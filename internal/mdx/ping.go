package mdx

func Ping() {
	isAlive := client.Ping()

	if isAlive {
		dp.Println("MangaDex API is alive")
	} else {
		dp.Println("MangaDex API is NOT alive")
	}
}
