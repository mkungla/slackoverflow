// Copyright © 2016 -2017 A-Frame authors.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package internal

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/howi-ce/howi/addon/application/plugin/cli"
	"github.com/howi-ce/howi/lib/filesystem/path"
	// Importing database driver
	_ "github.com/mattn/go-sqlite3"
)

const (
	stackExchangeQuestionSchema = `CREATE TABLE IF NOT EXISTS "StackExchangeQuestion" (
  "QID" INTEGER PRIMARY KEY,
  "UID" INTEGER,
  "title" TEXT,
  "creationDate" TIMESTAMP,
  "lastActivityDate" TIMESTAMP,
  "shareLink" TEXT,
  "closedReason" TEXT,
  "tags" TEXT,
  "site" TEXT,
  "isAnswered" INTEGER,
  "score" INTEGER,
  "viewCount" INTEGER,
  "answerCount" INTEGER,
  "commentCount" INTEGER,
  "upVoteCount" INTEGER,
  "downVoteCount" INTEGER,
  "deleteVoteCount" INTEGER,
  "favoriteCount" INTEGER,
  "reOpenVoteCount" INTEGER)`

	stackExchangeUserSchema = `CREATE TABLE IF NOT EXISTS "StackExchangeUser" (
  "UID" INTEGER PRIMARY KEY,
  "displayName" TEXT,
  "profileImage" TEXT,
  "link" TEXT,
  "reputation" INTEGER,
  "acceptRate" INTEGER,
  "badgeBronze" INTEGER,
  "badgeSilver" INTEGER,
  "badgeGold" INTEGER)`

	slackQuestionSchema = `CREATE TABLE IF NOT EXISTS "SlackQuestion" (
  "QID" INTEGER,
  "channel" TEXT,
  "ts" TEXT)`
)

// Database for SlackOverflow
type Database struct {
	db   *sql.DB
	file path.Obj
}

// SetPath to sqllite database file
func (d *Database) SetPath(file path.Obj) {
	d.file = file
}

// VerifyTables of SlackOverflow database
func (d *Database) VerifyTables(w *cli.Worker) (err error) {
	err = d.open()
	if err != nil {
		return err
	}
	_, err = d.db.Exec(stackExchangeQuestionSchema)
	if err != nil {
		return err
	}
	w.Log.Debug("DB: Stack Exchange Question Schema ok")

	_, err = d.db.Exec(stackExchangeUserSchema)
	if err != nil {
		return err
	}
	w.Log.Debug("DB: Stack Exchange User Schema ok")

	_, err = d.db.Exec(slackQuestionSchema)
	if err != nil {
		return err
	}
	w.Log.Debug("DB: Slack Question Schema ok")

	return nil
}

// LatestStackExchangeQuestion return latest locl question if any
func (d *Database) LatestStackExchangeQuestion() (StackExchangeQuestion, error) {
	q := StackExchangeQuestion{}
	err := d.open()
	if err != nil {
		return q, err
	}
	err = d.db.QueryRow("SELECT * FROM StackExchangeQuestion ORDER BY QID DESC LIMIT 1").Scan(
		&q.QID,
		&q.UID,
		&q.Title,
		&q.CreationDate,
		&q.LastActivityDate,
		&q.ShareLink,
		&q.ClosedReason,
		&q.Tags,
		&q.Site,
		&q.IsAnswered,
		&q.Score,
		&q.ViewCount,
		&q.AnswerCount,
		&q.CommentCount,
		&q.UpVoteCount,
		&q.DownVoteCount,
		&q.DeleteVoteCount,
		&q.FavoriteCount,
		&q.ReOpenVoteCount,
	)
	return q, err
}

// SyncStackExchangeUserShallowUser Create or update Stack Exchange User
func (d *Database) SyncStackExchangeUserShallowUser(user ShallowUserObj) (msg string, err error) {
	err = d.open()
	if err != nil {
		return msg, err
	}
	existingUser := d.FindStackExchangeUser(user.UID)
	// Map the user
	seu := StackExchangeUser{}
	seu.UID = user.UID
	seu.DisplayName = user.DisplayName
	seu.ProfileImage = user.ProfileImage
	seu.Link = user.Link
	seu.Reputation = user.Reputation
	seu.AcceptRate = user.AcceptRate
	seu.BadgeBronze = user.BadgeCounts.Bronze
	seu.BadgeSilver = user.BadgeCounts.Silver
	seu.BadgeGold = user.BadgeCounts.Gold

	// If there is no update needed
	if existingUser == seu {
		return "User: " + seu.DisplayName + " is already up to date", err
	}

	// Create new user or update existing one
	if existingUser.UID > 0 {
		msg, err = d.StackExchangeUserUpdate(seu)
	} else {
		msg, err = d.StackExchangeUserCreate(seu)
	}
	return msg, err
}

// SyncStackExchangeQuestion create or update question recieved from defined site
func (d *Database) SyncStackExchangeQuestion(q QuestionObj, site string) (msg string, err error) {

	err = d.open()
	if err != nil {
		return msg, err
	}

	existingQuestion := d.FindStackExchangeQuestion(q.QID)

	// Map the Question
	seq := StackExchangeQuestion{}
	seq.QID = q.QID
	seq.UID = q.Owner.UID
	seq.Title = q.Title
	seq.CreationDate = time.Unix(q.CreationDate, 0).UTC()
	seq.LastActivityDate = time.Unix(q.LastActivityDate, 0).UTC()
	seq.ShareLink = q.ShareLink
	seq.ClosedReason = q.ClosedReason
	seq.Tags = strings.Join(q.Tags, ";")
	seq.Site = site
	seq.IsAnswered = q.IsAnswered
	seq.Score = q.Score
	seq.ViewCount = q.ViewCount
	seq.AnswerCount = q.AnswerCount
	seq.CommentCount = q.CommentCount
	seq.UpVoteCount = q.UpVoteCount
	seq.DownVoteCount = q.DownVoteCount
	seq.DeleteVoteCount = q.DeleteVoteCount
	seq.FavoriteCount = q.FavoriteCount
	seq.ReOpenVoteCount = q.ReOpenVoteCount

	// If there is no update needed
	if existingQuestion == seq {
		return fmt.Sprintf("Question: %d is already up to date.", seq.QID), err
	}

	// Create new user or update existing one
	if existingQuestion.QID > 0 {
		msg, err = d.StackExchangeQuestionUpdate(seq)
	} else {
		msg, err = d.StackExchangeQuestionCreate(seq)
	}
	return msg, err
}

// FindStackExchangeQuestion by ID
func (d *Database) FindStackExchangeQuestion(QID int) StackExchangeQuestion {
	q := StackExchangeQuestion{}
	err := d.open()
	if err != nil {
		return q
	}
	stmt, err := d.db.Prepare(`SELECT * FROM StackExchangeQuestion WHERE QID = ?`)
	defer stmt.Close()
	if err != nil {
		return q
	}
	_ = stmt.QueryRow(QID).Scan(
		&q.QID,
		&q.UID,
		&q.Title,
		&q.CreationDate,
		&q.LastActivityDate,
		&q.ShareLink,
		&q.ClosedReason,
		&q.Tags,
		&q.Site,
		&q.IsAnswered,
		&q.Score,
		&q.ViewCount,
		&q.AnswerCount,
		&q.CommentCount,
		&q.UpVoteCount,
		&q.DownVoteCount,
		&q.DeleteVoteCount,
		&q.FavoriteCount,
		&q.ReOpenVoteCount,
	)
	return q
}

// StackExchangeQuestionCreate stores new Question
func (d *Database) StackExchangeQuestionCreate(seq StackExchangeQuestion) (msg string, err error) {
	err = d.open()
	if err != nil {
		return msg, err
	}
	tx, err := d.db.Begin()
	if err != nil {
		return msg, err
	}

	stmt, err := d.db.Prepare(`INSERT INTO StackExchangeQuestion
      (QID, UID, title, creationDate, lastActivityDate, shareLink, closedReason,
        tags, site, isAnswered, score, viewCount, answerCount, commentCount,
        upVoteCount, downVoteCount, deleteVoteCount, favoriteCount, reOpenVoteCount)
      VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19);`)
	defer stmt.Close()
	if err != nil {
		return msg, err
	}

	if _, err := stmt.Exec(
		seq.QID,
		seq.UID,
		seq.Title,
		seq.CreationDate,
		seq.LastActivityDate,
		seq.ShareLink,
		seq.ClosedReason,
		seq.Tags,
		seq.Site,
		seq.IsAnswered,
		seq.Score,
		seq.ViewCount,
		seq.AnswerCount,
		seq.CommentCount,
		seq.UpVoteCount,
		seq.DownVoteCount,
		seq.DeleteVoteCount,
		seq.FavoriteCount,
		seq.ReOpenVoteCount,
	); err != nil {
		tx.Rollback()
		return "Error had to Rollback question table", err
	}

	return fmt.Sprintf("Question: %d created.", seq.QID), nil

}

// StackExchangeQuestionUpdate updates existsing question
func (d *Database) StackExchangeQuestionUpdate(seq StackExchangeQuestion) (msg string, err error) {
	err = d.open()
	if err != nil {
		return msg, err
	}
	tx, err := d.db.Begin()
	if err != nil {
		return msg, err
	}
	stmt, err := d.db.Prepare(`UPDATE StackExchangeQuestion SET
    UID=?, title=?, creationDate=?, lastActivityDate=?, shareLink=?, closedReason=?,
      tags=?, site=?, isAnswered=?, score=?, viewCount=?, answerCount=?, commentCount=?,
      upVoteCount=?, downVoteCount=?, deleteVoteCount=?, favoriteCount=?, reOpenVoteCount=?
      WHERE QID=?;`)
	defer stmt.Close()
	if err != nil {
		return msg, err
	}

	if _, err := stmt.Exec(
		seq.UID,
		seq.Title,
		seq.CreationDate,
		seq.LastActivityDate,
		seq.ShareLink,
		seq.ClosedReason,
		seq.Tags,
		seq.Site,
		seq.IsAnswered,
		seq.Score,
		seq.ViewCount,
		seq.AnswerCount,
		seq.CommentCount,
		seq.UpVoteCount,
		seq.DownVoteCount,
		seq.DeleteVoteCount,
		seq.FavoriteCount,
		seq.ReOpenVoteCount,
		seq.QID,
	); err != nil {
		tx.Rollback()
		return "Error had to Rollback question update", err
	}

	return fmt.Sprintf("Question: %d updated.", seq.QID), nil
}

// FindStackExchangeUser by ID
func (d *Database) FindStackExchangeUser(UID int) StackExchangeUser {
	user := StackExchangeUser{}
	err := d.open()
	if err != nil {
		return user
	}
	stmt, err := d.db.Prepare(`SELECT * FROM StackExchangeUser WHERE UID = ?`)
	defer stmt.Close()
	if err != nil {
		return user
	}
	_ = stmt.QueryRow(UID).Scan(
		&user.UID,
		&user.DisplayName,
		&user.ProfileImage,
		&user.Link,
		&user.Reputation,
		&user.AcceptRate,
		&user.BadgeBronze,
		&user.BadgeSilver,
		&user.BadgeGold,
	)
	return user
}

// StackExchangeUserCreate creates new User
func (d *Database) StackExchangeUserCreate(seu StackExchangeUser) (msg string, err error) {
	err = d.open()
	if err != nil {
		return msg, err
	}
	tx, err := d.db.Begin()
	if err != nil {
		return msg, err
	}

	stmt, err := d.db.Prepare(`INSERT INTO StackExchangeUser
      (UID, displayName, profileImage, link, reputation, acceptRate, badgeBronze, badgeSilver, badgeGold)
      VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9);`)
	defer stmt.Close()
	if err != nil {
		return msg, err
	}

	if _, err := stmt.Exec(
		seu.UID,
		seu.DisplayName,
		seu.ProfileImage,
		seu.Link,
		seu.Reputation,
		seu.AcceptRate,
		seu.BadgeBronze,
		seu.BadgeSilver,
		seu.BadgeGold,
	); err != nil {
		tx.Rollback()
		return "Error had to Rollback", err
	}
	return "User " + seu.DisplayName + " created.", nil
}

// FindSlackQuestion link
func (d *Database) FindSlackQuestion(QID int) SlackQuestion {
	q := SlackQuestion{}
	err := d.open()
	if err != nil {
		return q
	}
	stmt, err := d.db.Prepare(`SELECT * FROM SlackQuestion WHERE QID = ?`)
	defer stmt.Close()
	if err != nil {
		return q
	}
	_ = stmt.QueryRow(QID).Scan(
		&q.QID,
		&q.Channel,
		&q.TS,
	)
	return q
}

// StackExchangeUserUpdate existsing User
func (d *Database) StackExchangeUserUpdate(seu StackExchangeUser) (msg string, err error) {
	err = d.open()
	if err != nil {
		return msg, err
	}
	tx, err := d.db.Begin()
	if err != nil {
		return msg, err
	}
	stmt, err := d.db.Prepare(`UPDATE StackExchangeUser SET
      displayName=?, profileImage=?, link=?, reputation=?, acceptRate=?, badgeBronze=?, badgeSilver=?, badgeGold=?
      WHERE UID=?;`)
	defer stmt.Close()
	if err != nil {
		return msg, err
	}

	if _, err := stmt.Exec(
		seu.DisplayName,
		seu.ProfileImage,
		seu.Link,
		seu.Reputation,
		seu.AcceptRate,
		seu.BadgeBronze,
		seu.BadgeSilver,
		seu.BadgeGold,
		seu.UID,
	); err != nil {
		tx.Rollback()
		return "Error had to Rollback", err
	}

	return "User :" + seu.DisplayName + " updated.", nil
}

// SlackQuestionCreate new Question
func (d *Database) SlackQuestionCreate(slq SlackQuestion) (msg string, err error) {
	err = d.open()
	if err != nil {
		return msg, err
	}
	tx, err := d.db.Begin()
	if err != nil {
		return msg, err
	}

	stmt, err := d.db.Prepare(`INSERT INTO SlackQuestion
      (QID, Channel, TS)
      VALUES($1,$2,$3);`)
	defer stmt.Close()
	if err != nil {
		return msg, err
	}

	if _, err := stmt.Exec(
		slq.QID,
		slq.Channel,
		slq.TS,
	); err != nil {
		tx.Rollback()
		return "Error had to Rollback question table", err
	}

	return fmt.Sprintf("Question link: %d created.", slq.QID), nil

}

// SlackQuestionDelete by ID
func (d *Database) SlackQuestionDelete(slq SlackQuestion) error {
	err := d.open()
	if err != nil {
		return err
	}
	stmt, err := d.db.Prepare(`DELETE FROM SlackQuestion WHERE QID = ?`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(slq.QID)
	defer stmt.Close()
	return err
}

// SlackQuestionGetAll linked questions
func (d *Database) SlackQuestionGetAll() (links []SlackQuestion, count int) {
	err := d.open()
	if err != nil {
		return
	}
	count = 0

	stmt, err := d.db.Prepare(`SELECT * FROM SlackQuestion ORDER BY ts DESC`)
	if err != nil {
		return
	}
	rows, err := stmt.Query()
	if err != nil {
		return links, count
	}
	defer stmt.Close()

	for rows.Next() {
		ql := SlackQuestion{}
		err = rows.Scan(
			&ql.QID,
			&ql.Channel,
			&ql.TS)
		if err != nil {
			log.Fatal(err)
		}
		links = append(links, ql)
		count++
	}

	return links, count
}

// StackExchangeQuestionTrackedIds return tracked question ids
func (d *Database) StackExchangeQuestionTrackedIds(qToWatch int) (ids string, count int) {
	err := d.open()
	if err != nil {
		return
	}
	count = 0
	ids = ""

	stmt, err := d.db.Prepare(`SELECT QID FROM StackExchangeQuestion ORDER BY creationDate DESC LIMIT ?`)
	defer stmt.Close()
	if err != nil {
		return ids, count
	}
	rows, err := stmt.Query(qToWatch)
	if err != nil {
		return ids, count
	}

	var idsMap []string
	for rows.Next() {
		var QID string
		err = rows.Scan(&QID)
		if err != nil {
			log.Fatal(err)
		}
		// in case of only one question asign that to idsMap
		ids = QID
		idsMap = append(idsMap, QID)
		count++
	}
	if count > 1 {
		ids = strings.Join(idsMap, ";")
	}

	return ids, count
}

// StackExchangeQuestionDelete by ID
func (d *Database) StackExchangeQuestionDelete(seq StackExchangeQuestion) error {
	err := d.open()
	if err != nil {
		return err
	}
	stmt, err := d.db.Prepare(`DELETE FROM StackExchangeQuestion WHERE QID = ?`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(seq.QID)
	defer stmt.Close()
	return err
}

// Close the database
func (d *Database) Close() error {
	return d.db.Close()
}

// StackExchangeQuestionsTracked returns tracked questions
func (d *Database) StackExchangeQuestionsTracked(qToWatch int) (questions []StackExchangeQuestion, count int) {
	err := d.open()
	if err != nil {
		return
	}
	count = 0

	stmt, err := d.db.Prepare(`SELECT * FROM StackExchangeQuestion ORDER BY creationDate DESC LIMIT ?`)
	defer stmt.Close()
	if err != nil {
		return questions, count
	}
	rows, err := stmt.Query(qToWatch)
	if err != nil {
		return questions, count
	}

	for rows.Next() {
		q := StackExchangeQuestion{}
		err = rows.Scan(
			&q.QID,
			&q.UID,
			&q.Title,
			&q.CreationDate,
			&q.LastActivityDate,
			&q.ShareLink,
			&q.ClosedReason,
			&q.Tags,
			&q.Site,
			&q.IsAnswered,
			&q.Score,
			&q.ViewCount,
			&q.AnswerCount,
			&q.CommentCount,
			&q.UpVoteCount,
			&q.DownVoteCount,
			&q.DeleteVoteCount,
			&q.FavoriteCount,
			&q.ReOpenVoteCount)
		if err != nil {
			log.Fatal(err)
		}
		// in case of only one question asign that to idsMap
		questions = append(questions, q)
		count++
	}

	return questions, count
}

// open database if it is not already open
func (d *Database) open() (err error) {
	if d.db != nil {
		return d.db.Ping()
	}
	d.db, err = sql.Open("sqlite3", d.file.Abs())
	return err
}

// SlackQuestion table
// Records in this table keep track of qustions between Stack Exchange and Slack
type SlackQuestion struct {
	// Id of StackExchangeQuestion question
	QID     int
	Channel string
	TS      string
}

// StackExchangeQuestion table
type StackExchangeQuestion struct {
	QID              int
	UID              int
	Title            string
	CreationDate     time.Time
	LastActivityDate time.Time
	ShareLink        string
	ClosedReason     string
	Tags             string
	Site             string
	IsAnswered       bool
	Score            int
	ViewCount        int
	AnswerCount      int
	CommentCount     int
	UpVoteCount      int
	DownVoteCount    int
	DeleteVoteCount  int
	FavoriteCount    int
	ReOpenVoteCount  int
}

// StackExchangeUser table
type StackExchangeUser struct {
	UID          int
	DisplayName  string
	ProfileImage string
	Link         string
	Reputation   int
	AcceptRate   int
	BadgeBronze  int
	BadgeSilver  int
	BadgeGold    int
}
