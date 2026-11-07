package httpapi
import "testing"
func TestCSVSafe(t *testing.T){cases:=map[string]string{"hello@example.com":"hello@example.com","=1+1":"'=1+1","@SUM(A1:A2)":"'@SUM(A1:A2)","-4":"'-4"};for in,want:=range cases{if got:=csvSafe(in);got!=want{t.Errorf("csvSafe(%q)=%q, want %q",in,got,want)}}}
func TestNewCode(t *testing.T){v,err:=newCode();if err!=nil{t.Fatal(err)};if len(v)!=10||v[:4]!="RLY-"{t.Fatalf("unexpected code %q",v)}}
