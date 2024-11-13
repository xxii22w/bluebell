package mysql

import (
	"bluebell/models"
	"strings"

	"github.com/jmoiron/sqlx"
)

// CreatePost 创建帖子
func CreatePost(p *models.Post) (err error) {
	sqlStr := `insert into post(
		post_id, title, content, author_id, community_id)
		values (?, ?, ?, ?, ?)
		`
	_, err = db.Exec(sqlStr, p.ID, p.Title, p.Content, p.AuthorID, p.CommunityID)
	return
}

// GetPostById 根据id查询单个帖子数据
func GetPostById(pid int64) (post *models.Post, err error) {
	post = new(models.Post)
	sqlStr := `select
	post_id, title, content, author_id, community_id, create_time
	from post
	where post_id = ?`
	err = db.Get(post, sqlStr, pid)
	return
}

// GetPostList 查询帖子列表函数 (限制每页贴子数)
func GetPostList(page, size int64) (posts []*models.Post, err error) {
	sqlStr := `select 
	post_id, title, content, author_id, community_id, create_time
	from post
	ORDER BY create_time
	DESC
	limit ?,?
	`
	posts = make([]*models.Post, 0, 2)
	err = db.Select(&posts, sqlStr, (page-1)*size, size)
	return
}

// GetPostListByIDs 根据给定的id列表查询帖子数据
func GetPostListByIDs(ids []string) (postList []*models.Post, err error) {
	sqlStr := `select post_id, title, content, author_id, community_id, create_time
	from post
	where post_id in (?)
	order by FIND_IN_SET(post_id, ?)`

	// 使用sqlx.In帮我们拼接语句和参数, 注意传入的参数是[]interface{} [1 2 3] => [1,2,3]
	query, args, err := sqlx.In(sqlStr, ids, strings.Join(ids, ","))
	if err != nil {
		return nil, err
	}
	// sqlx.In 返回带 `?` bindvar的查询语句, 我们使用Rebind()重新绑定
	query = db.Rebind(query)
	err = db.Select(&postList, query, args...)
	return
}
