package main

import (
	"fmt"
	"lab-system-backend/internal/database"
	"lab-system-backend/internal/models"
)

func main() {
	database.InitDB()

	// Clean up existing users and departments
	database.DB.Exec("DELETE FROM users")
	database.DB.Exec("DELETE FROM departments")

	// 1. Root Organization
	root := models.Department{Name: "新材料研究院", Type: "institute"}
	database.DB.Create(&root)
	fmt.Printf("Created Root: %s\n", root.Name)

	// 2. Teams and Department
	teams := []struct {
		Name  string
		Type  string
		Roles []struct {
			RoleName string
			Names    []string
		}
	}{
		{
			Name: "分析团队", Type: "team",
			Roles: []struct {
				RoleName string
				Names    []string
			}{
				{"team_leader", []string{"张伟"}},
				{"member", []string{"李娜", "王强", "赵敏", "陈志"}},
			},
		},
		{
			Name: "研发团队A", Type: "team",
			Roles: []struct {
				RoleName string
				Names    []string
			}{
				{"team_leader", []string{"刘洋"}},
				{"member", []string{"杨勇", "吴刚", "孙丽"}},
			},
		},
		{
			Name: "研发团队B", Type: "team",
			Roles: []struct {
				RoleName string
				Names    []string
			}{
				{"team_leader", []string{"黄勇"}},
				{"member", []string{"周杰", "徐强", "朱军"}},
			},
		},
		{
			Name: "研发团队C", Type: "team",
			Roles: []struct {
				RoleName string
				Names    []string
			}{
				{"team_leader", []string{"马超"}},
				{"member", []string{"何静", "罗平"}},
			},
		},
		{
			Name: "条件保障部", Type: "department",
			Roles: []struct {
				RoleName string
				Names    []string
			}{
				{"director", []string{"郑海"}},
				{"procurement_specialist", []string{"钱多多"}}, // 采购专员
				{"measurement_specialist", []string{"吴测量"}}, // 计量专员
				{"safety_specialist", []string{"安以轩"}},      // 安全专员
			},
		},
	}

	for _, t := range teams {
		dept := models.Department{
			Name:     t.Name,
			Type:     t.Type,
			ParentID: &root.ID,
		}
		database.DB.Create(&dept)
		fmt.Printf("Created Dept: %s\n", dept.Name)

		for _, roleGroup := range t.Roles {
			for _, name := range roleGroup.Names {
				user := models.User{
					Username:     generateUsername(roleGroup.RoleName, name),
					RealName:     name,
					DepartmentID: dept.ID,
					Role:         roleGroup.RoleName,
				}
				database.DB.Create(&user)
				fmt.Printf("  - Created User: %s (%s)\n", user.RealName, user.Role)
			}
		}
	}
}

func generateUsername(role, name string) string {
	// Simple username generation logic, e.g. pinyin or random
	// using name for now to be unique enough for demo
	return name // in real app use pinyin
}
