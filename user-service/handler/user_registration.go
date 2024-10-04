package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"user-service/db"
	"user-service/model"
	"user-service/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// hash password
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}
	user.Password = hashedPassword
	if len(user.ID) == 0 {
		user.ID = primitive.NewObjectID()
	}

	//check if user already exists
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}}, // 1 for ascending order
		Options: options.Index().SetUnique(true),
	}
	db.MI.DB.Collection("users").Indexes().CreateOne(context.TODO(), indexModel)

	// insert user into db
	_, err = db.MI.DB.Collection("users").InsertOne(c, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User already exists"})
		return
	}
	userJson, _ := json.Marshal(user)
	err = utils.EmitEvent("user", string(userJson))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error emitting event"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func HashPassword(password string) (string, error) {
	// hash password
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func GetUsers(c *gin.Context) {
	var users []model.User
	cursor, err := db.MI.DB.Collection("users").Find(context.Background(), bson.D{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching users"})
		return
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var user model.User
		cursor.Decode(&user)
		users = append(users, user)
	}
	c.JSON(http.StatusOK, users)
}

func GetUser(c *gin.Context) {
	var user model.User
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	err = db.MI.DB.Collection("users").FindOne(context.Background(), bson.D{{Key: "_id", Value: objectID}}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user"})
		return
	}
	c.JSON(http.StatusOK, user)
}
