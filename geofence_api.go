package main

import (
	"net/http"
	_ "github.com/lib/pq"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"gopkg.in/ini.v1"
	"fmt"
	"os"
	"github.com/gin-contrib/cors"
	"strings"
)

var DB *gorm.DB
var username string 
var password string
var database string
var sslmode string
var hostname string
var api_port string

var id string
var area_id string
var area_name string
var kd_kantor_hcms string
var kd_kantor_passion string

type Geofences struct {
  ID string `json:"id"`
  Area_id string `json:"news_id"`
  Area_name string `json:"area_name"`
  Kd_kantor_hcms string `json:"kd_kantor_hcms"`
  Kd_kantor_passion string `json:"kd_kantor_passion"`
}

const service_key string = "ed86c18a-50c2-4017-8bbe-733c0591477a"

func init() {
    var errDB error

    cfg, errCfg := ini.Load("config.ini")
    if errCfg != nil {
        fmt.Printf("Fail to read file: %v", errCfg)
        os.Exit(1)
    }
    if errCfg != nil {
        panic(errCfg)
    }

	hostname = cfg.Section("DATABASE").Key("hostname").String()
    username = cfg.Section("DATABASE").Key("username").String()
    password = cfg.Section("DATABASE").Key("password").String()
    password = strings.Replace(password, "[number-sign]", "#", -1)
    database = cfg.Section("DATABASE").Key("database").String()
    sslmode = cfg.Section("DATABASE").Key("sslmode").String()
    api_port = cfg.Section("APP").Key("api_port").String()

    var queryString = fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=%s", hostname, username, password, database, sslmode)

    DB, errDB = gorm.Open("postgres", queryString)
    DB.LogMode(true) 
    if errDB != nil {
        panic(errDB)
    }

    DB.AutoMigrate(&Geofences{})
}

func infoAPI(c *gin.Context) {
	message := "Version 1.0. This is a data web service for geofence areas. How to use this API can contact the developer."
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "description": message })
}

func checkArea(c *gin.Context) {
	latitude := c.Query("latitude")
	longitude := c.Query("longitude")
	key := c.Query("key")

	SearchQuery := `
	SELECT id, area_id, area_name, kd_kantor_hcms, kd_kantor_passion 
	FROM geofences as tb 
	WHERE ST_Intersects(ST_GeomFromGeoJSON(tb.computed), ST_SetSRID(ST_MakePoint(?, ?),4326)) = '1' 
	LIMIT 1
	`
	rows, err := DB.Raw(SearchQuery, longitude, latitude).Rows()
	defer rows.Close()

	var id string
	var area_id string
	var area_name string
	var kd_kantor_hcms string
	var kd_kantor_passion string

	for rows.Next() {
		rows.Scan(&id, &area_id, &area_name, &kd_kantor_hcms, &kd_kantor_passion)
	}
	
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": nil })
	} else {
		if len(id) < 1 {
			c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": nil })
		} else {
			if key == service_key {
				c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": Geofences{id, area_id, area_name, kd_kantor_hcms, kd_kantor_passion}  })
			} else {
				c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "Key not valid"})
			}
		}
		
	}	
}

func main() {
	fmt.Printf("Service starting at %s", api_port)
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(cors.Default())

	v1 := router.Group("/api/v1/geofences/")
	{
		v1.GET("/", infoAPI)
		v1.GET("/check_area", checkArea)
	}
	router.Run(api_port)
        fmt.Println("Stoped")
}
