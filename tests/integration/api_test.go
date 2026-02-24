package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"agent-hub/internal/model"
	"agent-hub/internal/testutil"
)

func doJSON(t *testing.T, r http.Handler, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()
	var buf *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, path, buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func decodeJSON(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	if rr.Body.Len() == 0 {
		return map[string]any{}
	}
	var m map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &m); err != nil {
		t.Fatalf("decode json: %v, body=%s", err, rr.Body.String())
	}
	return m
}

func asInt64(t *testing.T, v any) int64 {
	t.Helper()
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case json.Number:
		i, _ := n.Int64()
		return i
	default:
		t.Fatalf("not a number: %#v", v)
		return 0
	}
}

func TestAPI_FullFlow(t *testing.T) {
	app := testutil.NewMySQLTestApp(t)

	var community model.Community
	if err := app.DB.First(&community).Error; err != nil {
		t.Fatalf("load community: %v", err)
	}

	// health
	{
		rr := doJSON(t, app.Router, http.MethodGet, "/health", nil, "")
		if rr.Code != http.StatusOK {
			t.Fatalf("health status=%d body=%s", rr.Code, rr.Body.String())
		}
	}

	// register alice
	alice := struct {
		username string
		email    string
		password string
	}{username: "alice", email: "alice@example.com", password: "password123"}

	var aliceToken string
	var aliceAgentID int64
	{
		rr := doJSON(t, app.Router, http.MethodPost, "/api/v1/auth/register", map[string]any{
			"username": alice.username,
			"email":    alice.email,
			"password": alice.password,
		}, "")
		if rr.Code != http.StatusCreated {
			t.Fatalf("register alice status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeJSON(t, rr)
		aliceToken, _ = out["token"].(string)
		aliceAgentID = asInt64(t, out["agent_id"])
		if aliceToken == "" || aliceAgentID == 0 {
			t.Fatalf("register alice missing token/agent_id: %v", out)
		}
	}

	// login alice
	{
		rr := doJSON(t, app.Router, http.MethodPost, "/api/v1/auth/login", map[string]any{
			"email":    alice.email,
			"password": alice.password,
		}, "")
		if rr.Code != http.StatusOK {
			t.Fatalf("login alice status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeJSON(t, rr)
		if tok, _ := out["token"].(string); tok == "" {
			t.Fatalf("login alice token empty: %v", out)
		}
	}

	// create post by alice
	var postID int64
	{
		rr := doJSON(t, app.Router, http.MethodPost, "/api/v1/posts", map[string]any{
			"community_id": community.ID,
			"title":        "Hello Agent Hub",
			"content":      "This is my first post.",
		}, aliceToken)
		if rr.Code != http.StatusCreated {
			t.Fatalf("create post status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeJSON(t, rr)
		postID = asInt64(t, out["id"])
		if postID == 0 {
			t.Fatalf("create post id missing: %v", out)
		}
	}

	// list posts
	{
		rr := doJSON(t, app.Router, http.MethodGet, "/api/v1/posts?sort_by=new&limit=20&offset=0", nil, "")
		if rr.Code != http.StatusOK {
			t.Fatalf("list posts status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeJSON(t, rr)
		if asInt64(t, out["total"]) < 1 {
			t.Fatalf("list posts total unexpected: %v", out)
		}
	}

	// get post
	{
		rr := doJSON(t, app.Router, http.MethodGet, "/api/v1/posts/"+strconv.FormatInt(postID, 10), nil, "")
		if rr.Code != http.StatusOK {
			t.Fatalf("get post status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeJSON(t, rr)
		if asInt64(t, out["id"]) != postID {
			t.Fatalf("get post id mismatch: %v", out)
		}
	}

	// update post by owner
	{
		rr := doJSON(t, app.Router, http.MethodPut, "/api/v1/posts/"+strconv.FormatInt(postID, 10), map[string]any{
			"title":   "Hello Agent Hub (updated)",
			"content": "Updated content for my first post.",
		}, aliceToken)
		if rr.Code != http.StatusOK {
			t.Fatalf("update post status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeJSON(t, rr)
		if title, _ := out["title"].(string); title == "" {
			t.Fatalf("update post title missing: %v", out)
		}
	}

	// register bob
	bob := struct {
		username string
		email    string
		password string
	}{username: "bob", email: "bob@example.com", password: "password123"}

	var bobToken string
	var bobAgentID int64
	{
		rr := doJSON(t, app.Router, http.MethodPost, "/api/v1/auth/register", map[string]any{
			"username": bob.username,
			"email":    bob.email,
			"password": bob.password,
		}, "")
		if rr.Code != http.StatusCreated {
			t.Fatalf("register bob status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeJSON(t, rr)
		bobToken, _ = out["token"].(string)
		bobAgentID = asInt64(t, out["agent_id"])
		if bobToken == "" || bobAgentID == 0 {
			t.Fatalf("register bob missing token/agent_id: %v", out)
		}
	}

	// update post by non-owner should be 403
	{
		rr := doJSON(t, app.Router, http.MethodPut, "/api/v1/posts/"+strconv.FormatInt(postID, 10), map[string]any{
			"title": "hacked",
		}, bobToken)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("update post by non-owner status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeJSON(t, rr)
		if errObj, ok := out["error"].(map[string]any); !ok || errObj["code"] != "FORBIDDEN" {
			t.Fatalf("expected error FORBIDDEN: %v", out)
		}
	}

	// bob comments on alice post => alice gets notification
	var commentID int64
	{
		rr := doJSON(t, app.Router, http.MethodPost, "/api/v1/posts/"+strconv.FormatInt(postID, 10)+"/comments", map[string]any{
			"content": "Nice post, thanks for sharing this with everyone!",
		}, bobToken)
		if rr.Code != http.StatusCreated {
			t.Fatalf("create comment status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeJSON(t, rr)
		commentID = asInt64(t, out["id"])
		if commentID == 0 {
			t.Fatalf("create comment id missing: %v", out)
		}
	}

	// list comments
	{
		rr := doJSON(t, app.Router, http.MethodGet, "/api/v1/posts/"+strconv.FormatInt(postID, 10)+"/comments?limit=20&offset=0", nil, "")
		if rr.Code != http.StatusOK {
			t.Fatalf("list comments status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeJSON(t, rr)
		if asInt64(t, out["total"]) < 1 {
			t.Fatalf("list comments total unexpected: %v", out)
		}
	}

	// bob upvotes then downvotes the post
	{
		rr := doJSON(t, app.Router, http.MethodPost, "/api/v1/posts/"+strconv.FormatInt(postID, 10)+"/vote", map[string]any{
			"vote_type": 1,
		}, bobToken)
		if rr.Code != http.StatusOK {
			t.Fatalf("vote post up status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeJSON(t, rr)
		if asInt64(t, out["net_votes"]) != 1 {
			t.Fatalf("net_votes expected 1: %v", out)
		}

		rr = doJSON(t, app.Router, http.MethodPost, "/api/v1/posts/"+strconv.FormatInt(postID, 10)+"/vote", map[string]any{
			"vote_type": -1,
		}, bobToken)
		if rr.Code != http.StatusOK {
			t.Fatalf("vote post down status=%d body=%s", rr.Code, rr.Body.String())
		}
		out = decodeJSON(t, rr)
		if asInt64(t, out["net_votes"]) != -1 {
			t.Fatalf("net_votes expected -1: %v", out)
		}
	}

	// bob follows alice, cannot follow self
	{
		rr := doJSON(t, app.Router, http.MethodPost, "/api/v1/agents/"+url.PathEscape(alice.username)+"/follow", map[string]any{
			"follow": true,
		}, bobToken)
		if rr.Code != http.StatusOK {
			t.Fatalf("follow alice status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeJSON(t, rr)
		if asInt64(t, out["followers_count"]) != 1 {
			t.Fatalf("followers_count expected 1: %v", out)
		}

		rr = doJSON(t, app.Router, http.MethodPost, "/api/v1/agents/"+url.PathEscape(bob.username)+"/follow", map[string]any{
			"follow": true,
		}, bobToken)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("follow self status=%d body=%s", rr.Code, rr.Body.String())
		}
	}

	// search agents / posts
	{
		rr := doJSON(t, app.Router, http.MethodGet, "/api/v1/search?type=agents&q=ali&limit=20&offset=0", nil, "")
		if rr.Code != http.StatusOK {
			t.Fatalf("search agents status=%d body=%s", rr.Code, rr.Body.String())
		}
		rr = doJSON(t, app.Router, http.MethodGet, "/api/v1/search?type=posts&q=updated&limit=20&offset=0", nil, "")
		if rr.Code != http.StatusOK {
			t.Fatalf("search posts status=%d body=%s", rr.Code, rr.Body.String())
		}
	}

	// leaderboard
	{
		rr := doJSON(t, app.Router, http.MethodGet, "/api/v1/leaderboard?type=points&limit=10", nil, "")
		if rr.Code != http.StatusOK {
			t.Fatalf("leaderboard points status=%d body=%s", rr.Code, rr.Body.String())
		}
		rr = doJSON(t, app.Router, http.MethodGet, "/api/v1/leaderboard?type=influence&limit=10", nil, "")
		if rr.Code != http.StatusOK {
			t.Fatalf("leaderboard influence status=%d body=%s", rr.Code, rr.Body.String())
		}
		rr = doJSON(t, app.Router, http.MethodGet, "/api/v1/leaderboard?type=content&limit=10", nil, "")
		if rr.Code != http.StatusOK {
			t.Fatalf("leaderboard content status=%d body=%s", rr.Code, rr.Body.String())
		}
	}

	// notifications for alice (from comment + follow)
	var notifID int64
	{
		rr := doJSON(t, app.Router, http.MethodGet, "/api/v1/notifications?limit=20&offset=0", nil, aliceToken)
		if rr.Code != http.StatusOK {
			t.Fatalf("list notifications status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeJSON(t, rr)
		items, ok := out["notifications"].([]any)
		if !ok || len(items) == 0 {
			t.Fatalf("notifications empty: %v", out)
		}
		first, ok := items[0].(map[string]any)
		if !ok {
			t.Fatalf("notification item shape unexpected: %v", items[0])
		}
		notifID = asInt64(t, first["id"])
		if notifID == 0 {
			t.Fatalf("notification id missing: %v", first)
		}

		rr = doJSON(t, app.Router, http.MethodPatch, "/api/v1/notifications/"+strconv.FormatInt(notifID, 10)+"/read", nil, aliceToken)
		if rr.Code != http.StatusNoContent {
			t.Fatalf("mark read status=%d body=%s", rr.Code, rr.Body.String())
		}

		rr = doJSON(t, app.Router, http.MethodPost, "/api/v1/notifications/read-all", nil, aliceToken)
		if rr.Code != http.StatusNoContent {
			t.Fatalf("mark all read status=%d body=%s", rr.Code, rr.Body.String())
		}
	}

	// bob deletes own comment
	{
		rr := doJSON(t, app.Router, http.MethodDelete, "/api/v1/comments/"+strconv.FormatInt(commentID, 10), nil, bobToken)
		if rr.Code != http.StatusNoContent {
			t.Fatalf("delete comment status=%d body=%s", rr.Code, rr.Body.String())
		}
	}

	// alice deletes post
	{
		rr := doJSON(t, app.Router, http.MethodDelete, "/api/v1/posts/"+strconv.FormatInt(postID, 10), nil, aliceToken)
		if rr.Code != http.StatusNoContent {
			t.Fatalf("delete post status=%d body=%s", rr.Code, rr.Body.String())
		}

		rr = doJSON(t, app.Router, http.MethodGet, "/api/v1/posts/"+strconv.FormatInt(postID, 10), nil, "")
		if rr.Code != http.StatusNotFound {
			t.Fatalf("get deleted post status=%d body=%s", rr.Code, rr.Body.String())
		}
	}

	_ = aliceAgentID
	_ = bobAgentID
}

