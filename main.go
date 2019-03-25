package main

func main() {
	exitCallback := make(chan bool)

	repository := InitRepository()

	adminServer := newAdminServer(repository)

	adminServer.startListeningAdmin(exitCallback)
	startListeningPublic(exitCallback)

	<-exitCallback
}
