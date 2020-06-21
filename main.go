package main

func main() {
	a := App{}
	a.InitializeDBConnection()
	a.InitializeRoutes()
}
