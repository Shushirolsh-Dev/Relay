package httpapi

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/Shushirolsh-Dev/Relay/internal/store"
)
type Config struct{AdminToken string}
type API struct{store *store.Store;cfg Config;mux *http.ServeMux}
func New(db *store.Store,cfg Config)http.Handler{a:=&API{store:db,cfg:cfg,mux:http.NewServeMux()};a.routes();return headers(a.mux)}
func(a *API)routes(){
	a.mux.HandleFunc("GET /api/health",a.health)
	a.mux.HandleFunc("POST /api/waitlist",a.join)
	a.mux.HandleFunc("GET /api/admin/stats",a.admin(a.stats))
	a.mux.HandleFunc("GET /api/admin/export.csv",a.admin(a.export))
	a.mux.Handle("GET /",http.FileServer(http.Dir("web")))
	a.mux.HandleFunc("GET /admin",func(w http.ResponseWriter,r *http.Request){http.ServeFile(w,r,"web/admin.html")})
}
func(a *API)health(w http.ResponseWriter,r *http.Request){if err:=a.store.Health();err!=nil{jsonOut(w,503,map[string]string{"status":"unavailable"});return};jsonOut(w,200,map[string]string{"status":"ok","service":"relay"})}
type joinInput struct{Email string `json:"email"`;ReferralCode string `json:"referral_code"`}
func(a *API)join(w http.ResponseWriter,r *http.Request){
	r.Body=http.MaxBytesReader(w,r.Body,4096);var in joinInput
	d:=json.NewDecoder(r.Body);d.DisallowUnknownFields();if err:=d.Decode(&in);err!=nil{jsonOut(w,400,map[string]string{"error":"Send a valid email address."});return}
	email:=strings.ToLower(strings.TrimSpace(in.Email));parsed,err:=mail.ParseAddress(email)
	if len(email)>254||err!=nil||parsed.Address!=email||!strings.Contains(email,"."){jsonOut(w,400,map[string]string{"error":"Enter a valid email address."});return}
	ref:=strings.TrimSpace(in.ReferralCode);if len(ref)>32{ref=""}
	code,err:=newCode();if err!=nil{jsonOut(w,500,map[string]string{"error":"Could not create your invite code."});return}
	item,err:=a.store.Add(email,ref,code);if errors.Is(err,store.ErrDuplicate){jsonOut(w,409,map[string]string{"error":"That email is already on the list."});return}
	if err!=nil{log.Printf("signup: %v",err);jsonOut(w,500,map[string]string{"error":"Something went wrong. Try again."});return}
	position,err:=a.store.Position();if err!=nil{position=0}
	jsonOut(w,201,map[string]any{"message":"You're on the list.","referral_code":item.ReferralCode,"position":position})
}
func(a *API)admin(next http.HandlerFunc)http.HandlerFunc{return func(w http.ResponseWriter,r *http.Request){
	if a.cfg.AdminToken==""{jsonOut(w,503,map[string]string{"error":"Admin access is not configured."});return}
	raw:=r.Header.Get("Authorization");const prefix="Bearer ";if !strings.HasPrefix(raw,prefix){jsonOut(w,401,map[string]string{"error":"Unauthorized."});return}
	provided:=strings.TrimPrefix(raw,prefix);if len(provided)!=len(a.cfg.AdminToken)||subtle.ConstantTimeCompare([]byte(provided),[]byte(a.cfg.AdminToken))!=1{jsonOut(w,401,map[string]string{"error":"Unauthorized."});return};next(w,r)
}}
func(a *API)stats(w http.ResponseWriter,r *http.Request){v,err:=a.store.Stats();if err!=nil{log.Printf("stats: %v",err);jsonOut(w,500,map[string]string{"error":"Could not load stats."});return};jsonOut(w,200,v)}
func(a *API)export(w http.ResponseWriter,r *http.Request){
	rows,err:=a.store.Export();if err!=nil{jsonOut(w,500,map[string]string{"error":"Could not export signups."});return};defer rows.Close()
	w.Header().Set("Content-Type","text/csv; charset=utf-8");w.Header().Set("Content-Disposition",`attachment; filename="relay-waitlist.csv"`)
	c:=csv.NewWriter(w);_ = c.Write([]string{"email","referral_code","referred_by","created_at"})
	for rows.Next(){var email,code,ref string;var created time.Time;if err:=rows.Scan(&email,&code,&ref,&created);err!=nil{http.Error(w,"Export failed",500);return};_ = c.Write([]string{csvSafe(email),csvSafe(code),csvSafe(ref),created.Format(time.RFC3339)})};c.Flush()
}
func newCode()(string,error){b:=make([]byte,3);if _,err:=rand.Read(b);err!=nil{return "",err};return "RLY-"+strings.ToUpper(hex.EncodeToString(b)),nil}
func csvSafe(s string)string{if s!=""&&strings.ContainsRune("=+-@\t\r",rune(s[0])){return "'"+s};return s}
func jsonOut(w http.ResponseWriter,status int,v any){w.Header().Set("Content-Type","application/json; charset=utf-8");w.WriteHeader(status);_ = json.NewEncoder(w).Encode(v)}
func headers(next http.Handler)http.Handler{return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
	w.Header().Set("X-Content-Type-Options","nosniff");w.Header().Set("Referrer-Policy","strict-origin-when-cross-origin");w.Header().Set("X-Frame-Options","DENY")
	w.Header().Set("Content-Security-Policy","default-src 'self'; style-src 'self' https://fonts.googleapis.com; font-src https://fonts.gstatic.com; script-src 'self'; img-src 'self' data:; connect-src 'self'; base-uri 'self'; frame-ancestors 'none'")
	next.ServeHTTP(w,r)
})}
