package controller

import (
	"fmt"
	"lh/common"
	"lh/global"
	"lh/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CustomTime time.Time

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	layouts := []string{
		"2006-01-02T15:04",
		time.RFC3339,
		"2006-01-02 15:04",
		"2006-01-02",
	}
	var parsed time.Time
	var err error
	for _, layout := range layouts {
		parsed, err = time.Parse(layout, s)
		if err == nil {
			*ct = CustomTime(parsed)
			return nil
		}
	}
	return fmt.Errorf("invalid time format: %s", s)
}

func (ct CustomTime) Time() time.Time {
	return time.Time(ct)
}

// GetStudentList 获取学生列表
func GetStudentList(c *gin.Context) {
	db := global.DB
	var students []models.User
	if err := db.Model(&models.User{}).
		Where("Role = ?", "student").Find(&students).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "数据库查询失败",
		})
		return
	}
	studentResponse := make([]string, len(students))
	for i, student := range students {
		studentResponse[i] = strconv.FormatUint(uint64(student.ID), 10)
	}
	c.JSON(http.StatusOK, gin.H{
		"student_ids": studentResponse,
	})
}

// GetStudentListWithGroup 获取带分组的学生列表
func GetStudentListWithGroup(c *gin.Context) {
	db := common.GetDB()

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit
	query := db.Model(&models.User{}).Where("Role = ?", "student")
	var students []models.User
	var total int64
	query.Count(&total)
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&students).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "数据库查询失败",
		})
		return
	}
	studentIDs := make([]uint, len(students))
	for i, stu := range students {
		studentIDs[i] = stu.ID
	}
	// 查询学生与小组的关联关系
	type GroupStudent struct {
		UserID  uint
		GroupID uint
	}
	var relations []GroupStudent

	db.Table("group_students").
		Select("user_id, group_id").
		Where("user_id IN ?", studentIDs).
		Scan(&relations)

	// 构建学生ID到小组ID列表的映射
	groupMap := make(map[uint][]string)
	for _, r := range relations {
		groupIDStr := strconv.FormatUint(uint64(r.GroupID), 10)
		groupMap[r.UserID] = append(groupMap[r.UserID], groupIDStr)
	}
	response := make([]gin.H, len(students))
	for i, stu := range students {
		groupIDs, exists := groupMap[stu.ID]
		if !exists {
			groupIDs = []string{} // 确保返回空数组而不是null
		}
		response[i] = gin.H{
			"user_id":    stu.ID,
			"username":   stu.Name,
			"telephone":  stu.Telephone,
			"email":      stu.Email,
			"role":       stu.Role,
			"group_ids":  groupIDs,
			"created_at": stu.CreatedAt,
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": response,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
		"message": "学生列表获取成功",
	})
}

// CreateStudentGroup 创建学生分组
func CreateStudentGroup(c *gin.Context) {
	// 创建分组的请求结构体
	var req struct {
		GroupName  string   `json:"group_name" binding:"required"`
		StudentIDs []string `json:"student_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}
	db := global.DB
	studentIDs := make([]uint, 0, len(req.StudentIDs))
	for _, idStr := range req.StudentIDs {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "无效的学生ID: " + idStr,
			})
			return
		}
		studentIDs = append(studentIDs, uint(id))
	}
	// 验证所有学生是否存在且角色是学生
	var studentCount int64
	if err := db.Model(&models.User{}).
		Where("id IN ? AND role = ?", studentIDs, "student").
		Count(&studentCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库查询失败",
		})
		return
	}

	if int(studentCount) != len(studentIDs) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "部分学生不存在或不是学生角色",
		})
		return
	}
	// 创建分组
	newGroup := models.Group{
		Name: req.GroupName,
	}

	// 使用事务确保数据一致性
	err := db.Transaction(func(tx *gorm.DB) error {
		// 创建分组记录
		if err := tx.Create(&newGroup).Error; err != nil {
			return err
		}

		// 准备关联关系
		association := tx.Model(&newGroup).Association("Student")
		if err := association.Error; err != nil {
			return err
		}

		// 添加学生到分组
		var students []models.User
		if err := tx.Find(&students, studentIDs).Error; err != nil {
			return err
		}

		if err := association.Append(students); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建分组失败: " + err.Error(),
		})
		return
	}
	// 创建分组的响应结构体
	var GroupResponse struct {
		GroupID    string   `json:"group_id"`
		GroupName  string   `json:"group_name"`
		StudentIDs []string `json:"student_ids"`
	}
	GroupResponse.GroupID = strconv.FormatUint(uint64(newGroup.ID), 10)
	GroupResponse.GroupName = newGroup.Name
	GroupResponse.StudentIDs = req.StudentIDs
	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"data":    GroupResponse,
		"message": "创建学生分组成功",
	})
}

// GetStudentGroup 获取分组列表
func GetStudentGroup(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	db := global.DB

	// 查询分组总数
	var total int64
	db.Model(&models.Group{}).Count(&total)

	// 查询分组数据
	var groups []models.Group
	result := db.Offset(offset).Limit(limit).Find(&groups)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库查询失败",
		})
		return
	}

	// 收集分组ID
	groupIDs := make([]uint, len(groups))
	for i, group := range groups {
		groupIDs[i] = group.ID
	}

	// 查询分组与学生的关联关系
	type GroupStudent struct {
		GroupID uint
		UserID  uint
	}
	var relations []GroupStudent

	if len(groupIDs) > 0 {
		db.Table("group_students").
			Select("group_id, user_id").
			Where("group_id IN ?", groupIDs).
			Scan(&relations)
	}

	// 构建分组ID到学生ID列表的映射
	groupStudentMap := make(map[uint][]string)
	for _, r := range relations {
		studentID := strconv.FormatUint(uint64(r.UserID), 10)
		groupStudentMap[r.GroupID] = append(groupStudentMap[r.GroupID], studentID)
	}
	// 分组响应结构体
	type GroupResponse struct {
		GroupID      string   `json:"group_id"`
		GroupName    string   `json:"group_name"`
		StudentCount int      `json:"student_count"`
		StudentIDs   []string `json:"student_ids"`
	}
	response := make([]GroupResponse, len(groups))
	for i, group := range groups {
		studentIDs, exists := groupStudentMap[group.ID]
		if !exists {
			studentIDs = []string{} // 确保返回空数组而不是null
		}

		response[i] = GroupResponse{
			GroupID:      strconv.FormatUint(uint64(group.ID), 10),
			GroupName:    group.Name,
			StudentCount: len(studentIDs),
			StudentIDs:   studentIDs,
		}
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": response,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
		"message": "分组列表获取成功",
	})
}

// UpdateStudentGroup 更新分组情况
func UpdateStudentGroup(c *gin.Context) {
	var req struct {
		GroupName  string   `json:"group_name"`
		StudentIDs []string `json:"student_ids"`
	}
	db := global.DB
	groupIDstr := c.Param("group_id")
	groupID := common.StrToUint(groupIDstr)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var group models.Group
	if err := db.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "分组不存在"})
		return
	}
	if req.GroupName == "" && req.StudentIDs == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "至少提供一个更新字段(group_name 或 student_ids)",
		})
		return
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		// 更新分组名称（如果提供）
		if req.GroupName != "" {
			group.Name = req.GroupName
			if err := tx.Model(&group).Update("name", req.GroupName).Error; err != nil {
				return err
			}
		}

		// 更新学生列表（如果提供）
		if req.StudentIDs != nil {
			// 将学生ID从字符串转换为uint
			studentIDs := make([]uint, 0, len(req.StudentIDs))
			for _, idStr := range req.StudentIDs {
				id, err := strconv.ParseUint(idStr, 10, 64)
				if err != nil {
					return err
				}
				studentIDs = append(studentIDs, uint(id))
			}

			// 验证所有学生是否存在且角色是学生
			var studentCount int64
			if err := tx.Model(&models.User{}).
				Where("id IN ? AND role = ?", studentIDs, "student").
				Count(&studentCount).Error; err != nil {
				return err
			}

			if int(studentCount) != len(studentIDs) {
				return gorm.ErrRecordNotFound
			}

			// 修正点1: 先清空关联
			if err := tx.Model(&group).Association("Student").Clear(); err != nil {
				return err
			}

			// 修正点2: 使用Append逐个添加学生
			for _, id := range studentIDs {
				if err := tx.Model(&group).Association("Student").Append(&models.User{Model: gorm.Model{ID: id}}); err != nil {
					return err
				}
			}
		}

		// 刷新分组对象以获取最新更新时间
		if err := tx.First(&group, group.ID).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "部分学生不存在或不是学生角色",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "更新分组失败: " + err.Error(),
			})
		}
		return
	}
	type UpdateGroupResponse struct {
		GroupID    string   `json:"group_id"`
		GroupName  string   `json:"group_name"`
		StudentIDs []string `json:"student_ids"`
		UpdatedAt  string   `json:"updated_at"`
	}
	// 构建响应
	response := UpdateGroupResponse{
		GroupID:    groupIDstr,
		GroupName:  group.Name,
		StudentIDs: req.StudentIDs,
		UpdatedAt:  group.UpdatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    response,
		"message": "分组更新成功",
	})
}

// DeleteStudentGroup 删除学生分组
func DeleteStudentGroup(c *gin.Context) {
	// 从路径参数获取分组ID
	groupIDStr := c.Param("group_id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的分组ID",
		})
		return
	}

	db := global.DB

	// 检查分组是否存在
	var group models.Group
	result := db.First(&group, groupID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "分组不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "数据库查询失败",
			})
		}
		return
	}

	// 使用事务确保数据一致性
	err = db.Transaction(func(tx *gorm.DB) error {
		// 删除分组与学生之间的关联关系
		if err := tx.Model(&group).Association("Student").Clear(); err != nil {
			return err
		}

		// 删除分组（软删除）
		if err := tx.Delete(&group).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除分组失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "分组删除成功",
	})
}
