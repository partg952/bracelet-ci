package braceletdbservice

import (
	"bracelet-cicd/internal/bracelet-DB-service/parser"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)


func main() {
	r := gin.Default()
	r.POST("/event" ,  func(ctx *gin.Context) {
		body ,err := io.ReadAll(ctx.Request.Body)
		if err!=nil {
			ctx.JSON(http.StatusBadRequest , gin.H{"error" : "failed to read request body"})
			return
		}
		var rawEventInstance parser.RawEvent
		err = json.Unmarshal(body , &rawEventInstance)
		if err!=nil {
			log.Printf("Error occurred while unmarshalling for raw event : %v" , err)
		}


	})
	r.Run()
}