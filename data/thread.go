package data

import (
	"time"
)

type Thread struct {
	Id        int
	Uuid      string
	Topic     string
	UserId    int
	CreatedAt time.Time
}

type Post struct {
	Id        int
	Uuid      string
	Body      string
	UserId    int
	ThreadId  int
	CreatedAt time.Time
}

type Category struct {
	Title    string
	ThreadId int
}

type Rating struct {
	UserId   int
	ObjectId int
	Liked    bool
}

// format the CreatedAt date to display nicely on the screen
func (thread *Thread) CreatedAtDate() string {
	return thread.CreatedAt.Format("Jan 2, 2006 at 3:04pm")
}

func (post *Post) CreatedAtDate() string {
	return post.CreatedAt.Format("Jan 2, 2006 at 3:04pm")
}

// get the number of posts in a thread
func (thread *Thread) NumReplies() (count int) {
	db := db()
	defer db.Close()
	rows, err := db.Query("SELECT count(*) FROM posts where thread_id = $1", thread.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return
		}
	}
	rows.Close()
	return
}

func (thread *Thread) NumLikesThread() (count int) {
	db := db()
	defer db.Close()
	rows, err := db.Query("SELECT count(*) FROM thread_rating where thread_id = $1 and liked = 1", thread.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return
		}
	}
	rows.Close()
	return
}

func (thread *Thread) NumDislikesThread() (count int) {
	db := db()
	defer db.Close()
	rows, err := db.Query("SELECT count(*) FROM thread_rating where thread_id = $1 and liked = 0", thread.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return
		}
	}
	rows.Close()
	return
}

func (post *Post) NumLikesPost() (count int) {
	db := db()
	defer db.Close()
	rows, err := db.Query("SELECT count(*) FROM post_rating where post_id = $1 and liked = 1", post.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return
		}
	}
	rows.Close()
	return
}

func (post *Post) NumDislikesPost() (count int) {
	db := db()
	defer db.Close()
	rows, err := db.Query("SELECT count(*) FROM post_rating where post_id = $1 and liked = 0", post.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return
		}
	}
	rows.Close()
	return
}

// get posts to a thread
func (thread *Thread) Posts() (posts []Post, err error) {
	db := db()
	defer db.Close()
	rows, err := db.Query("SELECT id, uuid, body, user_id, thread_id, created_at FROM posts where thread_id = $1", thread.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		post := Post{}
		if err = rows.Scan(&post.Id, &post.Uuid, &post.Body, &post.UserId, &post.ThreadId, &post.CreatedAt); err != nil {
			return
		}
		posts = append(posts, post)
	}
	rows.Close()
	return
}

// Create a new thread
func (user *User) CreateThread(topic string, categories []string, imgName string) (conv Thread, err error) {
	db := db()
	defer db.Close()

	_, err = db.Exec("insert into threads (uuid, topic, user_id, created_at) values ($1, $2, $3, $4)", CreateUUID(), topic, user.Id, time.Now())
	if err != nil {
		return
	}

	err = db.QueryRow("SELECT id, uuid, topic, user_id, created_at FROM threads ORDER BY id DESC").Scan(&conv.Id, &conv.Uuid, &conv.Topic, &conv.UserId, &conv.CreatedAt)
	if err != nil {
		return
	}

	err = conv.ImageUpload(imgName)
	if err != nil {
		return
	}

	err = conv.AssignCategory(categories)
	if err != nil {
		return
	}
	return
}

func (thread *Thread) ImageUpload(imgName string) (err error) {
	if imgName == "" {
		return
	}
	db := db()
	defer db.Close()

	_, err = db.Exec("insert into images (thread_id, img_name) values ($1, $2)", thread.Id, imgName)
	if err != nil {
		return
	}
	return
}

func (thread *Thread) AssignCategory(categories []string) (err error) {
	db := db()
	defer db.Close()

	for i := range categories {
		_, err := db.Exec("insert into categories (title, thread_id) values ($1, $2)", categories[i], thread.Id)
		if err != nil {
			return err
		}
	}
	return
}

func (thread *Thread) DeleteThread() (err error) {
	db := db()
	defer db.Close()

	_, err = db.Exec("delete from threads where uuid = $1", thread.Uuid)
	if err != nil {
		return err
	}
	return
}

// Create a new post to a thread
func (user *User) CreatePost(conv Thread, body string) (post Post, err error) {
	db := db()
	defer db.Close()

	_, err = db.Exec("insert into posts (uuid, body, user_id, thread_id, created_at) values ($1, $2, $3, $4, $5)", CreateUUID(), body, user.Id, conv.Id, time.Now())
	if err != nil {
		return
	}

	err = db.QueryRow("SELECT id, uuid, body, user_id, thread_id, created_at FROM posts").Scan(&post.Id, &post.Uuid, &post.Body, &post.UserId, &post.ThreadId, &post.CreatedAt)
	if err != nil {
		return
	}
	return
}

func (user *User) RateThread(conv Thread) (err error) {
	db := db()
	defer db.Close()

	rating := Rating{}
	err = db.QueryRow("select * from thread_rating where user_id = $1 and thread_id = $2", user.Id, conv.Id).Scan(&rating.UserId, &rating.ObjectId, &rating.Liked)

	if rating != (Rating{}) {

		if rating.Liked == true {
			user.DeleteThreadRating(conv)
			return
		} else {
			_, err = db.Exec("update thread_rating set liked = 1 where user_id = $1 and thread_id = $2", rating.UserId, rating.ObjectId)
			if err != nil {
				return
			}
			return
		}
	}
	_, err = db.Exec("insert into thread_rating (user_id, thread_id, liked) values ($1, $2, $3)", user.Id, conv.Id, true)
	if err != nil {
		return
	}
	return
}

func (user *User) UnrateThread(conv Thread) (err error) {
	db := db()
	defer db.Close()
	rating := Rating{}
	db.QueryRow("select * from thread_rating where user_id = $1 and thread_id = $2", user.Id, conv.Id).Scan(&rating.UserId, &rating.ObjectId, &rating.Liked)
	if rating != (Rating{}) {
		if rating.Liked == false {
			user.DeleteThreadRating(conv)
			return
		} else {
			_, err = db.Exec("update thread_rating set liked = 0 where user_id = $1 and thread_id = $2", rating.UserId, rating.ObjectId)
			if err != nil {
				return
			}
			return
		}
	}
	_, err = db.Exec("insert into thread_rating (user_id, thread_id, liked) values ($1, $2, $3)", user.Id, conv.Id, false)
	if err != nil {
		return
	}
	return
}

func (user *User) DeleteThreadRating(conv Thread) (err error) {
	db := db()
	defer db.Close()
	_, err = db.Exec("delete from thread_rating where user_id = $1 and thread_id = $2", user.Id, conv.Id)
	if err != nil {
		return
	}
	return
}

func (user *User) RatePost(post Post) (err error) {
	db := db()
	defer db.Close()

	rating := Rating{}
	err = db.QueryRow("select * from post_rating where user_id = $1 and post_id = $2", user.Id, post.Id).Scan(&rating.UserId, &rating.ObjectId, &rating.Liked)
	if rating != (Rating{}) {
		if rating.Liked == true {
			user.DeletePostRating(post)
			return
		} else {
			_, err = db.Exec("update post_rating set liked = 1 where user_id = $1 and post_id = $2", rating.UserId, rating.ObjectId)
			if err != nil {
				return
			}
			return
		}
	}
	_, err = db.Exec("insert into post_rating (user_id, post_id, liked) values ($1, $2, $3)", user.Id, post.Id, true)
	if err != nil {
		return
	}
	return
}

func (user *User) UnratePost(post Post) (err error) {
	db := db()
	defer db.Close()
	rating := Rating{}
	db.QueryRow("select * from post_rating where user_id = $1 and post_id = $2", user.Id, post.Id).Scan(&rating.UserId, &rating.ObjectId, &rating.Liked)
	if rating != (Rating{}) {
		if rating.Liked == false {
			user.DeletePostRating(post)
			return
		} else {
			_, err = db.Exec("update post_rating set liked = 0 where user_id = $1 and post_id = $2", rating.UserId, rating.ObjectId)
			if err != nil {
				return
			}
			return
		}
	}
	_, err = db.Exec("insert into post_rating (user_id, post_id, liked) values ($1, $2, $3)", user.Id, post.Id, false)
	if err != nil {
		return
	}
	return
}

func (user *User) DeletePostRating(post Post) (err error) {
	db := db()
	defer db.Close()
	_, err = db.Exec("delete from post_rating where user_id = $1 and post_id = $2", user.Id, post.Id)
	if err != nil {
		return
	}
	return
}

// Get all threads in the database and returns it
func Threads() (threads []Thread, err error) {
	db := db()
	defer db.Close()
	rows, err := db.Query("SELECT id, uuid, topic, user_id, created_at FROM threads ORDER BY created_at DESC")
	if err != nil {
		return
	}
	for rows.Next() {
		conv := Thread{}
		if err = rows.Scan(&conv.Id, &conv.Uuid, &conv.Topic, &conv.UserId, &conv.CreatedAt); err != nil {
			return
		}
		threads = append(threads, conv)
	}
	rows.Close()
	return
}

// Get a thread by the UUID
func ThreadByUUID(uuid string) (conv Thread, err error) {
	db := db()
	defer db.Close()
	conv = Thread{}
	err = db.QueryRow("SELECT id, uuid, topic, user_id, created_at FROM threads WHERE uuid = $1", uuid).
		Scan(&conv.Id, &conv.Uuid, &conv.Topic, &conv.UserId, &conv.CreatedAt)
	return
}

func PostByUUID(uuid string) (post Post, err error) {
	db := db()
	defer db.Close()
	post = Post{}
	err = db.QueryRow("SELECT id, uuid, body, user_id, thread_id, created_at FROM posts WHERE uuid = $1", uuid).
		Scan(&post.Id, &post.Uuid, &post.Body, &post.UserId, &post.ThreadId, &post.CreatedAt)
	return
}

func (post *Post) ThreadUUIDbyPostUUID() (uuid string) {
	db := db()
	defer db.Close()
	db.QueryRow("SELECT uuid FROM threads WHERE id = $1", post.ThreadId).
		Scan(&uuid)
	return
}

// Get the user who started this thread
func (thread *Thread) User() (user User) {
	db := db()
	defer db.Close()
	user = User{}
	db.QueryRow("SELECT id, uuid, name, email, created_at FROM users WHERE id = $1", thread.UserId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt)
	return
}

// Get the user who wrote the post
func (post *Post) User() (user User) {
	db := db()
	defer db.Close()
	user = User{}
	db.QueryRow("SELECT id, uuid, name, email, created_at FROM users WHERE id = $1", post.UserId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt)
	return
}

func (user *User) CreatedThreads() (threads []Thread, err error) {
	db := db()
	defer db.Close()
	rows, err := db.Query("select id, uuid, topic, user_id, created_at from threads where user_id = $1 order by created_at DESC", user.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		conv := Thread{}
		if err = rows.Scan(&conv.Id, &conv.Uuid, &conv.Topic, &conv.UserId, &conv.CreatedAt); err != nil {
			return
		}
		threads = append(threads, conv)
	}
	rows.Close()
	return
}

func (user *User) LikedThreads() (threads []Thread, err error) {
	db := db()
	defer db.Close()
	rows, err := db.Query("select thread_id from thread_rating where user_id = $1 and liked = 1", user.Id)
	if err != nil {
		return
	}
	var threadIds []int
	for rows.Next() {
		var id int
		if err = rows.Scan(&id); err != nil {
			return
		}
		threadIds = append(threadIds, id)
	}
	rows.Close()

	for _, id := range threadIds {
		conv := Thread{}

		err = db.QueryRow("select id, uuid, topic, user_id, created_at from threads where id = $1 order by created_at DESC", id).
			Scan(&conv.Id, &conv.Uuid, &conv.Topic, &conv.UserId, &conv.CreatedAt)
		if err != nil {
			return
		}
		threads = append(threads, conv)
	}
	if err != nil {
		return
	}
	return
}

func ThreadsByCategories(categories []string) (threads []Thread, err error) {
	db := db()
	defer db.Close()

	var threadIds []int
	for _, title := range categories {
		rows, err := db.Query("select thread_id from categories where title = $1", title)
		if err != nil {
			return []Thread{}, err
		}
		for rows.Next() {
			var id int
			if err = rows.Scan(&id); err != nil {
				return []Thread{}, err
			}
			threadIds = append(threadIds, id)
		}
		rows.Close()
	}

	var temp []int
	for i := 0; i < len(threadIds); i++ {
		exists := false
		for j := i + 1; j < len(threadIds); j++ {
			if threadIds[i] == threadIds[j] {
				exists = true
			}
		}
		if !exists {
			temp = append(temp, threadIds[i])
		}
	}
	threadIds = temp
	for _, id := range threadIds {
		conv := Thread{}
		err = db.QueryRow("select id, uuid, topic, user_id, created_at from threads where id = $1 order by created_at DESC", id).
			Scan(&conv.Id, &conv.Uuid, &conv.Topic, &conv.UserId, &conv.CreatedAt)
		if err != nil {
			return
		}
		threads = append(threads, conv)
	}
	if err != nil {
		return
	}
	return
}

func (thread *Thread) Categories() (categories string) {
	db := db()
	defer db.Close()
	rows, err := db.Query("select title from categories where thread_id = $1", thread.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		var title string
		if err = rows.Scan(&title); err != nil {
			return
		}
		categories += title + " | "
	}
	return
}

func (thread *Thread) ImageNameByThread() (imgName string) {
	db := db()
	defer db.Close()

	err := db.QueryRow("SELECT img_name FROM images where thread_id = $1", thread.Id).Scan(&imgName)
	if err != nil {
		return
	}
	return
}
