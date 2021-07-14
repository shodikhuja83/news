package db

import (
	"github.com/mailru/dbr"
)

type SubscribeModel struct {
	UserId int64 `db:"user_id"`
	Tag string `db:"tag"`
	ReadenArticles []string `db:"readen_articles"`
}

func InsertNewSubscribeToSubscribes(s *dbr.Session, model SubscribeModel) (err error) {
	_, err = s.InsertInto("subscribes").
		Columns("user_id", "tag").
		Record(&model).
		Exec()
	return
}

func DeleteSubscribeFromSubscribes(s *dbr.Session, userId int64, tag string) (err error) {
	_, err = s.DeleteFrom("subscribes").
		Where("user_id = ? AND tag = ?", userId, tag).
		Exec()
	return
}

func SelectAllSubscribes(s *dbr.Session) (models []SubscribeModel, err error) {
	_, err = s.Select("user_id", "tag").
		From("subscribes").
		LoadStructs(&models)
	return
}

func AddReadenArticleToSubscribe(s *dbr.Session, userId int64, tag, article string) (err error) {
	_, err = s.Update("subscribes").
		Set("readen_articles", dbr.Expr(`array_append("readen_articles", '` + article + `')`)).
		Where("user_id = ? AND tag = ?", userId, tag).
		Exec()
	return
}

func IsSubscribeContainsArticle(s *dbr.Session, userId int64, tag, article string) (contains bool, err error) {
	selectStr := "readen_articles @> ARRAY['" + article + "']"
	err = s.Select(selectStr).
		From("subscribes").
		Where("user_id = ? AND tag = ?", userId, tag).
		LoadValue(&contains)
	return
}

func SelectUsersSubscribeTags(s *dbr.Session, userId int64) (subscribedTags []string, err error) {
	_, err = s.Select("tag").
		From("subscribes").
		Where("user_id = ?", userId).
		LoadValues(&subscribedTags)
	return
}