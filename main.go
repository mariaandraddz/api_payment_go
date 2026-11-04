package main

func main() {
	router := gin.Default();
	router.GET("/health", healthcheck.Health)

	router.Run(":8080");
}
