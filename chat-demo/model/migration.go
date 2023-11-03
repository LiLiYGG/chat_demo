package model

// 迁移函数
// 创建表时添加后缀
func migration() {
	DB.Set("gorm:table_options", "charset=utf8mb4").AutoMigrate(&User{})
}
