package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	_ "modernc.org/sqlite"
)

var ErrDuplicate=errors.New("email already registered")
type Signup struct { ID int64 `json:"id"`; Email string `json:"email"`; ReferralCode string `json:"referral_code"`; ReferredBy string `json:"referred_by,omitempty"`; CreatedAt time.Time `json:"created_at"` }
type RefStat struct { Code string `json:"code"`; Count int `json:"count"` }
type Stats struct { Total int `json:"total"`; Today int `json:"today"`; Recent []Signup `json:"recent"`; TopSources []RefStat `json:"top_referrers"` }
type Store struct { db *sql.DB }

func Open(path string)(*Store,error){
	db,err:=sql.Open("sqlite",path); if err!=nil{return nil,err}; db.SetMaxOpenConns(1)
	s:=&Store{db:db}
	_,err=db.Exec(`PRAGMA journal_mode=WAL;
CREATE TABLE IF NOT EXISTS signups(id INTEGER PRIMARY KEY AUTOINCREMENT,email TEXT NOT NULL UNIQUE,referral_code TEXT NOT NULL UNIQUE,referred_by TEXT NOT NULL DEFAULT '',created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);
CREATE INDEX IF NOT EXISTS idx_signups_created_at ON signups(created_at);
CREATE INDEX IF NOT EXISTS idx_signups_referred_by ON signups(referred_by);`)
	if err!=nil{db.Close();return nil,fmt.Errorf("initialize schema: %w",err)}; return s,nil
}
func(s *Store)Close()error{return s.db.Close()}
func(s *Store)Add(email,ref,code string)(Signup,error){
	res,err:=s.db.Exec("INSERT INTO signups(email,referral_code,referred_by) VALUES(?,?,?)",email,code,ref)
	if err!=nil{if strings.Contains(strings.ToLower(err.Error()),"unique"){return Signup{},ErrDuplicate};return Signup{},err}
	id,err:=res.LastInsertId();if err!=nil{return Signup{},err};var v Signup
	err=s.db.QueryRow("SELECT id,email,referral_code,referred_by,created_at FROM signups WHERE id=?",id).Scan(&v.ID,&v.Email,&v.ReferralCode,&v.ReferredBy,&v.CreatedAt);return v,err
}
func(s *Store)Position()(int,error){var n int;err:=s.db.QueryRow("SELECT COUNT(*) FROM signups").Scan(&n);return n,err}
func(s *Store)Health()error{return s.db.Ping()}
func(s *Store)Stats()(Stats,error){
	var v Stats
	if err:=s.db.QueryRow("SELECT COUNT(*) FROM signups").Scan(&v.Total);err!=nil{return v,err}
	if err:=s.db.QueryRow("SELECT COUNT(*) FROM signups WHERE date(created_at)=date('now')").Scan(&v.Today);err!=nil{return v,err}
	rows,err:=s.db.Query("SELECT id,email,referral_code,referred_by,created_at FROM signups ORDER BY created_at DESC LIMIT 12");if err!=nil{return v,err}
	for rows.Next(){var x Signup;if err:=rows.Scan(&x.ID,&x.Email,&x.ReferralCode,&x.ReferredBy,&x.CreatedAt);err!=nil{rows.Close();return v,err};v.Recent=append(v.Recent,x)}
	if err:=rows.Err();err!=nil{rows.Close();return v,err};rows.Close()
	refs,err:=s.db.Query("SELECT referred_by,COUNT(*) FROM signups WHERE referred_by<>'' GROUP BY referred_by ORDER BY COUNT(*) DESC LIMIT 5");if err!=nil{return v,err}
	defer refs.Close();for refs.Next(){var x RefStat;if err:=refs.Scan(&x.Code,&x.Count);err!=nil{return v,err};v.TopSources=append(v.TopSources,x)}
	return v,refs.Err()
}
func(s *Store)Export()(*sql.Rows,error){return s.db.Query("SELECT email,referral_code,referred_by,created_at FROM signups ORDER BY created_at DESC")}
