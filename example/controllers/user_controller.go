package controllers

import (
	"fmt"
	"time"

	"hertz-controller/framework/controller"
)

type UserController struct {
	controller.BaseController
}

type User struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Active      bool     `json:"active"`
	CreatedAt   string   `json:"created_at"`
	LastLogin   string   `json:"last_login"`
	Permissions []string `json:"permissions"`
}

func (c *UserController) GetIndex() {
	page := c.GetInt("page", 1)
	size := c.GetInt("size", 10)
	fmt.Println("当前页:", page, "每页大小:", size)

	// 模拟用户数据
	users := []User{
		{
			ID:          1,
			Name:        "张三",
			Email:       "zhangsan@example.com",
			Active:      true,
			CreatedAt:   "2024-01-15 10:30:00",
			LastLogin:   "2024-07-29 09:15:00",
			Permissions: []string{"read", "write"},
		},
		{
			ID:          2,
			Name:        "李四",
			Email:       "lisi@example.com",
			Active:      true,
			CreatedAt:   "2024-02-20 14:20:00",
			LastLogin:   "2024-07-28 16:45:00",
			Permissions: []string{"read"},
		},
		{
			ID:          3,
			Name:        "王五",
			Email:       "wangwu@example.com",
			Active:      false,
			CreatedAt:   "2024-03-10 09:45:00",
			LastLogin:   "2024-07-25 11:30:00",
			Permissions: []string{"read", "write", "admin"},
		},
	}

	// 模拟分页数据
	pagination := map[string]any{
		"HasPrev":  page > 1,
		"HasNext":  page < 3,
		"PrevPage": page - 1,
		"NextPage": page + 1,
		"Pages": []map[string]any{
			{"Page": 1, "IsCurrent": page == 1},
			{"Page": 2, "IsCurrent": page == 2},
			{"Page": 3, "IsCurrent": page == 3},
		},
	}

	c.SetData("Title", "用户管理")
	c.SetData("Users", users)
	c.SetData("Pagination", pagination)
	c.Render("user/index.html")
}

func (c *UserController) GetInfo() {
	userId := c.GetInt("id", 1)

	// 模拟根据ID获取用户
	user := User{
		ID:          userId,
		Name:        "张三",
		Email:       "zhangsan@example.com",
		Active:      true,
		CreatedAt:   "2024-01-15 10:30:00",
		LastLogin:   "2024-07-29 09:15:00",
		Permissions: []string{"read", "write", "user:create", "user:edit"},
	}

	c.SetData("Title", "用户详情")
	c.SetData("User", user)
	c.Render("user/info.html")
}

func (c *UserController) PostCreate() {
	name := c.GetForm("name")
	email := c.GetForm("email")
	password := c.GetForm("password")

	if name == "" || email == "" || password == "" {
		c.JSON(map[string]any{
			"success": false,
			"message": "请填写完整信息",
		})
		return
	}

	// 这里应该是保存到数据库的逻辑
	newUser := User{
		ID:          4,
		Name:        name,
		Email:       email,
		Active:      true,
		CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
		LastLogin:   "",
		Permissions: []string{"read"},
	}

	c.JSON(map[string]any{
		"success": true,
		"message": "用户创建成功",
		"user":    newUser,
	})
}

func (c *UserController) PutUpdate() {
	userId := c.GetString("id")
	name := c.GetForm("name")
	email := c.GetForm("email")

	if userId == "" {
		c.JSON(map[string]any{
			"success": false,
			"message": "用户ID不能为空",
		})
		return
	}

	c.JSON(map[string]any{
		"success": true,
		"message": "用户更新成功",
		"data": map[string]any{
			"id":    userId,
			"name":  name,
			"email": email,
		},
	})
}

func (c *UserController) DeleteRemove() {
	userId := c.GetString("id")

	if userId == "" {
		c.JSON(map[string]any{
			"success": false,
			"message": "用户ID不能为空",
		})
		return
	}

	// 这里应该是从数据库删除的逻辑
	c.JSON(map[string]any{
		"success": true,
		"message": "用户删除成功",
		"id":      userId,
	})
}
