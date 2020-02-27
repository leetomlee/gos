package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	_ "gopkg.in/mgo.v2/bson"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Book struct {
	Id       string `bson:"_id"`
	BookName string `bson:"book_name" json:"Name"`
	Category string `json:"CName"`
	Author   string `bson:"author"`
	UTime    string `bson:"u_time" json:"UTime"`
	BookDesc string `bson:"book_desc" json:"Desc,omitempty"`
	//Status      string    `bson:"status"`
	Cover         string `bson:"cover" json:"Img"`
	LastChapter   string `bson:"last_chapter" `
	LastChapterId string `bson:"last_chapter_id"`
}
type User struct {
	Id       bson.ObjectId `bson:"_id" json:"id,omitempty"`
	Name     string        `form:"name" bson:"name" json:"name" binding:"required"`
	PassWord string        `form:"password" bson:"password" json:"password,omitempty" binding:"required"`
	EMail    string        `form:"email" bson:"email" json:"email"`
}
type LoginUser struct {
	Name     string        `form:"name" bson:"name" json:"name" binding:"required"`
	PassWord string        `form:"password" bson:"password" json:"password,omitempty" binding:"required"`
}
type RegUser struct {
	Name     string `form:"name" bson:"name" json:"name" binding:"required"`
	PassWord string `form:"password" bson:"password" json:"password" binding:"required"`
	EMail    string `form:"email" bson:"email" json:"email" binding:"required"`
}
type BookDetail struct {
	Book            Book
	SameAuthorBooks []Book
}

/**
根据category 分页books
*/
type CateBook struct {
	BookName string `bson:"book_name" json:"bookName"`
	Author   string `json:"author"`
	Cover    string `json:"cover"`
	Id       string `bson:"_id" json:"id"`
}
type Chapter struct {
	ChapterName string `bson:"chapter_name" json:"chapterName"`
	ChapterId   string `bson:"_id" json:"chapterId"`
}
type BookContent struct {
	Id      string `bson:"_id" json:"id"`
	Content string `bson:"content" json:"content"`
}
type Resutlt struct {
	Id string `bson:"_id"' `
}

type Account struct {
	Id   bson.ObjectId `bson:"_id"`
	Name string
	IdS  []string `bson:"ids"`
}

func main() {
	key := "not so bad"

	gin.DisableConsoleColor()
	//f, _ := os.Create("gin.log")
	//gin.DefaultWriter = io.MultiWriter(f)
	ignoresPath := []string{"/login", "/register"}
	//gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(func(c *gin.Context) {

		path := c.FullPath()
		i := 0
		f := true
		for ; i < len(ignoresPath); i++ {
			if ignoresPath[i] == path {
				f = false
				break
			}
		}
		if f {
			auth, e := c.Cookie("auth")
			if e != nil {
				c.JSON(401, gin.H{"code": 401, "msg": "需要认证", "data": ""})
				c.Abort()
				return
			}
			decodeString, e := base64.StdEncoding.DecodeString(auth)
			if e != nil {
				c.Abort()
				return
			}
			if strings.Split(string(decodeString), ":")[1] != key {
				c.Abort()
				return
			}
		}
	})
	//db, err := gorm.Open("mysql", "root:lx123456zx@tcp(120.27.244.128:3306)/book?charset=utf8&parseTime=true")
	//db.SingularTable(true)
	//defer db.Close()
	//if err != nil {
	//	panic(err)
	//}
	/*
		mongodb define
	*/
	session, err := mgo.Dial("120.27.244.128")
	session.SetMode(mgo.Monotonic, true)

	if err != nil {
		panic(err)
	}
	defer session.Close()
	bookDB := session.DB("book").C("book")
	accountDB := session.DB("book").C("account")
	chapterDB := session.DB("book").C("chapter")
	r.PATCH("/password", func(c *gin.Context) {
		var user User
		if err := c.ShouldBind(&user); err != nil {
			panic(err)
		}
		count, err2 := accountDB.Find(bson.M{"name": user.Name, "email": user.EMail}).Count()
		if err2 != nil {
			panic(err2)
		}
		if count > 0 {
			accountDB.Update(bson.M{"name": user.Name}, bson.M{"$set": bson.M{"password": user.PassWord}})
			c.JSON(http.StatusOK, gin.H{"msg": "修改密码成功", "code": 200, "data": ""})
		}

	})
	r.POST("/register", func(c *gin.Context) {
		var regUser RegUser
		if err := c.ShouldBind(&regUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "需要认证参数", "code": 400, "data": ""})
			return
		}
		n, e := accountDB.Find(bson.M{"name": regUser.Name}).Count()
		if e != nil {
			return
		}
		if n > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "用户已存在", "code": 400, "data": ""})
			return
		}
		/**
		simple password encry
		*/
		hash := sha1.New()
		io.WriteString(hash, regUser.PassWord)
		regUser.PassWord = string(hash.Sum(nil))
		e1 := accountDB.Insert(regUser)
		if e1 != nil {
			return
		}
		c.JSON(http.StatusOK, gin.H{"msg": "注册成功", "code": 200, "data": ""})

	})
	r.POST("/login", func(c *gin.Context) {
		var user LoginUser
		if err := c.ShouldBind(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "需要认证参数", "code": 400, "data": ""})
			return
		}
		var users []User
		h := sha1.New()
		io.WriteString(h, user.PassWord)
		password := string(h.Sum(nil))
		if e := accountDB.Find(bson.M{"name": user.Name, "password": password}).All(&users); e != nil {
			panic(e)
		}
		if len(users) >= 1 {
			auth := bytes.NewBufferString(users[0].Id.Hex() + ":" + key)
			users[0].Id = ""
			users[0].PassWord = ""

			c.SetCookie("auth", base64.StdEncoding.EncodeToString(auth.Bytes()), -1, "/", "120.27.244.128", false, true)
			c.JSON(http.StatusOK, gin.H{"msg": "登陆成功", "code": 200, "data": users[0]})
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": "查无此人", "code": 400, "data": ""})
		}

	})
	books := r.Group("/book")
	{
		books.GET("", func(c *gin.Context) {
			var bookDetail BookDetail
			id := c.Query("id")
			var book Book
			if e := bookDB.FindId(id).One(&book); e != nil {
				panic(e)
			}
			bookDetail.Book = book
			var bks []Book
			m := []bson.M{
				{"$match": bson.M{"_id": bson.M{"$ne": book.Id}, "author": book.Author}},
				{"$project": bson.M{"book_desc": 0}},
			}
			if e := bookDB.Pipe(m).All(&bks); e != nil {
				panic(e)
			}
			bookDetail.SameAuthorBooks = bks
			c.JSON(http.StatusOK, gin.H{"msg": "", "code": 200, "data": bookDetail})

		})
		books.GET("/shelf", func(c *gin.Context) {
			account := getAccountFromCookie(c, accountDB)
			var books []Book
			if account.IdS != nil {
				if e := bookDB.Find(bson.M{"_id": bson.M{"$in": account.IdS}}).All(&books); e != nil {
					panic(e)
				}
			}
			c.JSON(http.StatusOK, gin.H{"msg": "", "code": "200", "data": books})
		})
		books.POST("", func(c *gin.Context) {
			bookId := c.PostForm("bookId")
			action := c.PostForm("action")
			account := getAccountFromCookie(c, accountDB)
			if account.IdS != nil {
				i := 0
				f := true
				for ; i < len(account.IdS); i++ {
					if account.IdS[i] == bookId {
						f = false

						break
					}
				}
				if f {
					if action == "add" {
						account.IdS = append(account.IdS, bookId)
						accountDB.UpdateId(account.Id, bson.M{"$set": bson.M{"ids": account.IdS}})
					}
				} else {
					if action == "del" {
						accountDB.UpdateId(account.Id, bson.M{"$set": bson.M{"ids": append(account.IdS[:i], account.IdS[i+1:]...)}})
					}
				}
			} else {
				if action == "add" {
					accountDB.UpsertId(account.Id, bson.M{"$set": bson.M{"ids": []string{bookId}}})
				}
			}

		})
		books.GET("/cates", func(c *gin.Context) {
			var data []Resutlt

			m := []bson.M{
				{"$group": bson.M{"_id": "$category", "count": bson.M{"$sum": 1}}},
				{"$sort": bson.M{"count": -1}},
				{"$project": bson.M{"count": 0, "_id": 1}},
			}
			bookDB.Pipe(m).All(&data)
			var tem []string
			for _, v := range data {
				tem = append(tem, v.Id)
			}
			c.JSON(200, gin.H{
				"data": tem,
				"code": 200,
				"msg":  "",
			})

		})
		books.GET("/cate/:cate/:page/:size", func(c *gin.Context) {
			page, e1 := strconv.Atoi(c.Param("page"))
			cate := c.Param("cate")
			size, e2 := strconv.Atoi(c.Param("size"))

			if e1 != nil {
				panic(e1)
			}
			if e2 != nil {
				panic(e2)
			}
			var datas []CateBook
			//db.Table("book").Where("category=?", cate).Select("author,book_name, cover,id").Offset((page - 1) * size).Limit(size).Scan(&datas)
			m := []bson.M{
				{"$match": bson.M{"category": cate}},
				{"$project": bson.M{"author": 1, "book_name": 1, "cover": 1, "_id": 1}},
				{"$skip": (page - 1) * size},
				{"$limit": size},
			}
			bookDB.Pipe(m).All(&datas)
			c.JSON(200, gin.H{
				"code": 200,
				"msg":  "",
				"data": datas,
			})
		})
		books.GET("/chapters/:id", func(c *gin.Context) {
			id := c.Param("id")
			var chapters []Chapter
			//db.Table("chapter").Where("book_id=?", id).Select("chapter_id,chapter_name").Order("chapter_id asc").Scan(&chapters)
			m := []bson.M{
				{"$match": bson.M{"book_id": id}},

				{"$project": bson.M{"chapter_name": 1, "_id": 1}},
				{"$sort": bson.M{"_id": 1}},
			}
			chapterDB.Pipe(m).All(&chapters)
			c.JSON(200, gin.H{
				"code": 200,
				"msg":  "",
				"data": chapters,
			})

		})
		books.GET("/chapter/:id", func(c *gin.Context) {
			id := c.Param("id")
			var result BookContent
			err = chapterDB.Find(bson.M{"_id": id}).One(&result)
			if err != nil {
				panic(err)
			}
			c.JSON(200, gin.H{
				"code": 200,
				"msg":  "",
				"data": result,
			})

		})
		books.GET("/search", func(c *gin.Context) {
			key := c.Query("key")
			var bks []Book
			//m := []bson.M{
			//	{"author": bson.M{"$regex": key, "$options": "$im"}},
			//{"$project": bson.M{"book_name": 1, "_id": 1}},
			//}
			var query []bson.M
			//var all []bson.M
			q1 := bson.M{"book_name": bson.M{"$regex": key, "$options": "$i$m"}}
			query = append(query, q1)
			q2 := bson.M{"author": bson.M{"$regex": key, "$options": "$i$m"}}
			//
			query = append(query, q2)

			//q3 := bson.M{"$project": bson.M{"_id": 1, "book_name": 1, "category": 1, "author": 1, "book_desc": 1, "cover": 1}}

			//all = append(all, q3)
			//db.Table("book").Where("book_name LIKE ?", key).Or("author like ?", key).Find(&bks)
			//book.Pipe(m).All(&bks)
			bookDB.Find(bson.M{"$or": query}).All(&bks)
			c.JSON(200, gin.H{
				"code": 200,
				"msg":  "",
				"data": bks,
			})
		})
		books.GET("/statistics", func(c *gin.Context) {
			//db.Table("chapter").Count(&chapters)
			n, e := chapterDB.Count()
			if e != nil {
				panic(e)
			}
			c.JSON(200, gin.H{
				"code": 200,
				"msg":  "",
				"data": strconv.Itoa(n),
			})

		})
		//books.GET("/chapters/sync", func(c *gin.Context) {
		//	command := exec.Command("go", "version")
		//	var buffer bytes.Buffer
		//
		//	command.Stdout = &buffer
		//
		//	run := command.Run()
		//	if run != nil {
		//		panic(run)
		//	}
		//	c.JSON(200, gin.H{
		//		"code": 200,
		//		"msg":  "",
		//		"data": buffer.String(),
		//	})
		//
		//})
	}
	r.Run("0.0.0.0:8082")
}
func getAccountFromCookie(c *gin.Context, accountDB *mgo.Collection) Account {
	cookie, e := c.Cookie("auth")
	if e != nil {
		panic(e)
	}
	deSrec, e := base64.StdEncoding.DecodeString(cookie)
	if e != nil {
		panic(e)
	}
	id := strings.Split(string(deSrec), ":")[0]
	var account Account

	if e := accountDB.FindId(bson.ObjectIdHex(id)).One(&account); e != nil {
		panic(e)
	}
	return account

}
