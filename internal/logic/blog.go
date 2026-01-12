package logic

// TODO: Relook a this struct when blogs are chosen. I am concerned about using the image porperty here. May delete later.
type Post struct {
	Title   string
	Author  string
	Subject string
	Image   string
}

func GetMockPosts() []Post {
	return []Post{
		{Title: "First Post", Author: "David", Subject: "Golang", Image: "/static/img1.png"},
	}
}
