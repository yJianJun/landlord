package router

import (
	"landlord/controllers"
	"landlord/service"
	"net/http"
)

func init() {
	http.HandleFunc("/", controllers.Index)
	http.HandleFunc("/login", controllers.Login)
	http.HandleFunc("/loginOut", controllers.LoginOut)
	http.HandleFunc("/reg", controllers.Register)

	http.HandleFunc("/ws", service.ServeWs)

	// 设置静态目录
	static := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", static))
}
