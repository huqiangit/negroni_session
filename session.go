package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"
)

//Session a library for negroni
type Session struct {
	clients map[string]struct {
		done <-chan time.Time
		flag int
	}
	mu sync.RWMutex
}

func (session *Session) isInclude(key string) bool {
	if _, ok := session.clients[key]; ok {
		return true
	}
	return false
}
func (session *Session) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	session.mu.RLock()
	for _, v := range r.Cookies() {
		if _, ok := session.clients[v.Value]; ok {
			session.mu.RUnlock()
			next(w, r)
			return
		}
	}
	session.mu.RUnlock()

	session.newSession(w, r, 60*time.Second)
	next(w, r)
}

func Add(w http.ResponseWriter, r *http.Request, durationSecond int64) {
	DefaultSession.newSession(w, r, time.Duration(durationSecond)*time.Second)
}
func IsInclude(w http.ResponseWriter, r *http.Request) bool {
	sessionkey, err := r.Cookie("key")
	if err != nil {
		return false
	}

	flag := DefaultSession.isInclude(sessionkey.Value)
	fmt.Println("negroni session include:", flag)
	fmt.Println("list:", DefaultSession.clients)
	return flag
}

func (session *Session) newSession(w http.ResponseWriter, r *http.Request, t time.Duration) {
	key := ""
	for {
		buffer := make([]byte, 40)
		rand.Read(buffer)
		key = hex.EncodeToString(buffer)
		session.mu.RLock()
		if _, ok := session.clients[key]; !ok {
			session.mu.RUnlock()
			break
		}
		session.mu.RUnlock()
	}

	cookie := &http.Cookie{
		Name:     "key",
		Value:    key,
		Path:     "/",
		HttpOnly: false,
		MaxAge:   int(t) / int(time.Second),
	}
	http.SetCookie(w, cookie)

	session.mu.Lock()
	session.clients[key] = struct {
		done <-chan time.Time
		flag int
	}{
		done: time.After(t),
	}
	session.mu.Unlock()

	fmt.Println("[session] new key(", time.Now(), "):", key)
	go func() {
		<-session.clients[key].done
		session.mu.Lock()
		delete(session.clients, key)
		session.mu.Unlock()
		fmt.Println("[session] del key(", time.Now(), "):", key)
	}()
}

//DefaultSession you should always use this one
var DefaultSession = &Session{
	clients: make(map[string]struct {
		done <-chan time.Time
		flag int
	}),
}
