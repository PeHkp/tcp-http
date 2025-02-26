package routes

func Routes(route string) string {

	switch route {
	case "/":
		return "Welcome to the home page"
	case "/about":
		return "Welcome to the about page"
	case "/contact":
		return "Welcome to the contact page"
	default:
		return "404 Not Found"
	}
}
