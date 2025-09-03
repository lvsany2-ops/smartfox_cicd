package common

import (
	"lh/models"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// jwt加密密钥
func getJWTKey() []byte {
	key := os.Getenv("JWT_SECRET")
	if key == "" {
		return []byte("a_secret_crect") // 默认密钥，生产环境不建议使用默认值
	}
	return []byte(key)
}

// Claims token的claim
type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.StandardClaims
}

// 发放token
func ReleaseToken(user models.User) (string, error) {

	//token的有效期
	expirationTime := time.Now().Add(7 * 24 * time.Hour)

	claims := &Claims{

		//自定义字段
		UserID: user.ID,
		Role:   user.Role,
		//标准字段
		StandardClaims: jwt.StandardClaims{

			//过期时间
			ExpiresAt: expirationTime.Unix(),
			//发放的时间
			IssuedAt: time.Now().Unix(),
			//发放者
			Issuer: "127.0.0.1",
			//主题
			Subject: "user token",
		},
	}

	//使用jwt密钥生成token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(getJWTKey())

	if err != nil {
		return "", err
	}

	//返回token
	return tokenString, nil
}
