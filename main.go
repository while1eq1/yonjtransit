package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"

	"github.com/ChimeraCoder/anaconda"
	_ "github.com/mattn/go-sqlite3"
)

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
		os.Exit(-1)
	}
}

func getStatusURL(tweetName string, tweetId int64) string {
	tweetId_s := strconv.FormatInt(tweetId, 10)
	url := "https://twitter.com/" + tweetName + "/status/" + tweetId_s
	return url
}

// TODO: create function to write data to logfile.
func writeLog() {
}

func getLastTweetId(db *sql.DB) int64 {
	var tweetId int64

	stmt := "select max(tweetID) from transitdb;"
	rows, err := db.Query(stmt)
	checkErr(err, "db.Query() failed: "+stmt)
	defer db.Close()

	for rows.Next() {
		rows.Scan(&tweetId)
	}

	return tweetId
}

func insertRec(db *sql.DB, tweetLog map[string][]string) bool {

	lastTweet := strconv.FormatInt(getLastTweetId(db), 10)

	for k, v := range tweetLog {
		if k > lastTweet {
			fmt.Printf("Inserting into db : %s %s %s %s %s\n", k, v[0], v[1], v[2], v[3])
			_, err := db.Exec("INSERT INTO transitdb (tweetId, timestamp, transitLine, url, yod) VALUES (?, ?, ?, ?, ?);", k, v[0], v[1], v[2], v[3])
			checkErr(err, "db.Exec() fatal!")
		}
	}
	fmt.Println("No new twitter feeds :(")
	return true
}

type ApiKeys struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
}

func main() {

	twtapikeys := ApiKeys{
		os.Getenv("TWITTER_CONSUMERKEY"),
		os.Getenv("TWITTER_CONSUMERSECRET"),
		os.Getenv("TWITTER_ACCESSTOKEN"),
		os.Getenv("TWITTER_ACCESSTOKENSECRET"),
	}

	anaconda.SetConsumerKey(twtapikeys.ConsumerKey)
	anaconda.SetConsumerSecret(twtapikeys.ConsumerSecret)
	api := anaconda.NewTwitterApi(twtapikeys.AccessToken, twtapikeys.AccessTokenSecret)

	username := "NJTRANSIT_ME"

	v := url.Values{}
	v.Set("count", "20")
	v.Set("screen_name", username)

	tweets, err := api.GetUserTimeline(v)
	checkErr(err, "api.GetUserTimeline() failed check connection or credentials")

	// Open the sqlite db
	db, err := sql.Open("sqlite3", "./db/transit.db")
	checkErr(err, "sql.Open() failed!")
	defer db.Close()

	tweetLog := make(map[string][]string) // Create a map to store our returned results
	for _, tweet := range tweets {
		url := getStatusURL(username, tweet.Id)
		tweetId := strconv.FormatInt(tweet.Id, 10)
		tweetLog[tweetId] = []string{
			tweet.CreatedAt,
			username,
			url,
			"1",
		}
	}

	insertRec(db, tweetLog)
	// fmt.Println(getLastTweetId(db))
}
