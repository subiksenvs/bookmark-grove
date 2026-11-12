package main

import (
 "encoding/json"
 "fmt"
 "io"
 "log"
 "net/http"
 "os"
 "path/filepath"
 "strings"
 "sync"
 "time"
)
type Link struct { ID string `json:"id"`; URL string `json:"url"`; Title string `json:"title"`; Note string `json:"note"`; Tags []string `json:"tags"`; Created time.Time `json:"created"` }
type Store struct { mu sync.Mutex; path string; Items []Link `json:"items"` }
func (s *Store) load(){ b,e:=os.ReadFile(s.path); if e==nil { _=json.Unmarshal(b,&s.Items) } }
func (s *Store) save() error { b,e:=json.MarshalIndent(s.Items,"","  "); if e!=nil{return e}; if e=os.MkdirAll(filepath.Dir(s.path),0755);e!=nil{return e}; return os.WriteFile(s.path,b,0600) }
func main(){ store:=&Store{path:"data/library.json"}; store.load(); mux:=http.NewServeMux()
 mux.Handle("/",http.FileServer(http.Dir("public")))
 mux.HandleFunc("/api/links",func(w http.ResponseWriter,r *http.Request){ w.Header().Set("Content-Type","application/json"); store.mu.Lock(); defer store.mu.Unlock(); switch r.Method {
 case "GET": q:=strings.ToLower(r.URL.Query().Get("q")); out:=[]Link{}; for _,x:=range store.Items { if q==""||strings.Contains(strings.ToLower(x.URL+" "+x.Title+" "+x.Note+" "+strings.Join(x.Tags," ")),q){out=append(out,x)} }; _=json.NewEncoder(w).Encode(out)
 case "POST": var x Link; if json.NewDecoder(http.MaxBytesReader(w,r.Body,1<<20)).Decode(&x)!=nil||!strings.HasPrefix(x.URL,"http://")&&!strings.HasPrefix(x.URL,"https://"){ http.Error(w,"Valid http(s) URL required",400);return }; if x.Title==""{x.Title=x.URL}; x.ID=fmt.Sprintf("%d",time.Now().UnixNano());x.Created=time.Now().UTC();store.Items=append([]Link{x},store.Items...);if e:=store.save();e!=nil{http.Error(w,e.Error(),500);return};w.WriteHeader(201);_=json.NewEncoder(w).Encode(x)
 default: http.Error(w,"Method not allowed",405)
 } })
 mux.HandleFunc("/api/links/",func(w http.ResponseWriter,r *http.Request){if r.Method!="DELETE"{http.Error(w,"Method not allowed",405);return};id:=strings.TrimPrefix(r.URL.Path,"/api/links/");store.mu.Lock();defer store.mu.Unlock();for i,x:=range store.Items{if x.ID==id{store.Items=append(store.Items[:i],store.Items[i+1:]...);if e:=store.save();e!=nil{http.Error(w,e.Error(),500);return};w.WriteHeader(204);return}};http.Error(w,"Not found",404)})
 mux.HandleFunc("/api/summary",func(w http.ResponseWriter,r *http.Request){store.mu.Lock();defer store.mu.Unlock(); tags:=map[string]int{};for _,x:=range store.Items{for _,t:=range x.Tags{tags[t]++}};w.Header().Set("Content-Type","application/json");_=json.NewEncoder(w).Encode(map[string]any{"total":len(store.Items),"tags":tags})})
 mux.HandleFunc("/api/coach",func(w http.ResponseWriter,r *http.Request){var in struct{Prompt string `json:"prompt"`};_=json.NewDecoder(r.Body).Decode(&in);answer:="Demo library coach: Capture the source, add a short note explaining why it matters, and tag it with the question you expect to revisit.";mode:="demo";if key:=os.Getenv("OPENAI_API_KEY");key!="" { payload,_:=json.Marshal(map[string]any{"model":env("OPENAI_MODEL","gpt-4.1-mini"),"input":"You are a careful knowledge-management coach. Give concise practical advice. User: "+in.Prompt});req,_:=http.NewRequest("POST","https://api.openai.com/v1/responses",strings.NewReader(string(payload)));req.Header.Set("Authorization","Bearer "+key);req.Header.Set("Content-Type","application/json");res,e:=http.DefaultClient.Do(req);if e==nil{defer res.Body.Close();b,_:=io.ReadAll(res.Body);if res.StatusCode<300{var o struct{Output []struct{Content []struct{Text string `json:"text"`} `json:"content"`} `json:"output"`};if json.Unmarshal(b,&o)==nil&&len(o.Output)>0&&len(o.Output[0].Content)>0{answer=o.Output[0].Content[0].Text;mode="ai"}}else{http.Error(w,"AI provider returned an error",502);return}} };w.Header().Set("Content-Type","application/json");_=json.NewEncoder(w).Encode(map[string]string{"answer":answer,"mode":mode})})
 port:=env("PORT","8792");log.Printf("Bookmark Grove: http://127.0.0.1:%s",port);log.Fatal(http.ListenAndServe("127.0.0.1:"+port,mux))
}
func env(k,d string)string{if x:=os.Getenv(k);x!=""{return x};return d}
