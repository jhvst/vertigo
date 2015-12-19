package routes

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	. "github.com/9uuso/vertigo/databases/sqlx"
	. "github.com/9uuso/vertigo/session"
	"github.com/9uuso/vertigo/render"

	"github.com/gorilla/context"
	"github.com/husobee/vestigo"
	"github.com/pborman/uuid"
)

func GetUser(r *http.Request) (User, error) {
	rv, ok := context.GetOk(r, "user")
	if !ok {
		return User{}, errors.New("context not set")
	}
	return rv.(User), nil
}

// CreateUser is a route which creates a new user struct according to posted parameters.
// Requires session cookie.
// Returns created user struct for API requests and redirects to "/user" on frontend ones.
func CreateUser(w http.ResponseWriter, r *http.Request) {

	user, err := GetUser(r)
	if err != nil {
		log.Println("route CreateUser, user context:", err)
		render.R.JSON(w, 500, map[string]interface{}{"error": "Internal server error"})
		return
	}

	if Settings.AllowRegistrations == false {
		log.Println("Denied a new registration.")
		switch Root(r) {
		case "api":
			render.R.JSON(w, 403, map[string]interface{}{"error": "New registrations are not allowed at this time."})
			return
		case "user":
			render.R.HTML(w, 403, "user/login", "New registrations are not allowed at this time.")
			return
		}
	}
	user, err = user.Insert()
	if err != nil {
		log.Println("route CreateUser, user.Insert:", err)
		if err.Error() == "user email exists" {
			render.R.JSON(w, 422, map[string]interface{}{"error": "Email already in use"})
			return
		}
		if err.Error() == "user location invalid" {
			render.R.JSON(w, 422, map[string]interface{}{"error": "Location invalid. Please use IANA timezone database compatible locations."})
			return
		}
		render.R.JSON(w, 500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	user, err = user.Login()
	if err != nil {
		log.Println("route CreateUser, user.Login:", err)
		render.R.JSON(w, 500, map[string]interface{}{"error": "Internal server error"})
		return
	}

	SessionSetValue(w, r, "id", user.ID)

	switch Root(r) {
	case "api":
		render.R.JSON(w, 200, user)
		return
	case "user":
		http.Redirect(w, r, "/user", 302)
		return
	}
}

// DeleteUser is a route which deletes a user from database according to session cookie.
// The function calls Login function inside, so it also requires password in POST data.
// Currently unavailable function on both API and frontend side.
// func DeleteUser(req *http.Request, user User) {
// 	user, err := user.Login()
// 	if err != nil {
// 		log.Println(err)
// 		render.R.JSON(50map[string]interface{}{"error": "Internal server error"})
// 		return
// 	}
// 	err = user.Delete(s)
// 	if err != nil {
// 		log.Println(err)
// 		render.R.JSON(50map[string]interface{}{"error": "Internal server error"})
// 		return
// 	}
// 	switch Root(r) {
// 	case "api":
// 		s.Delete("user")
// 		render.R.JSON(20map[string]interface{}{"status": "User successfully deleted"})
// 		return
// 	case "user":
// 		s.Delete("user")
// 		render.R.HTML(20"User successfully deleted", nil)
// 		return
// 	}
// 	render.R.JSON(50map[string]interface{}{"error": "Internal server error"})
// }

// ReadUser is a route which fetches user according to parameter "id" on API side and according to retrieved
// session cookie on frontend side.
// Returns user struct with all posts merged to object on API call. Frontend call will render user "home" page, "user/index.tmpl".
func ReadUser(w http.ResponseWriter, r *http.Request) {
	var user User
	switch Root(r) {
	case "api":
		id, err := strconv.Atoi(vestigo.Param(r, "id"))
		if err != nil {
			log.Println("route ReadUser, strconv.Atoi:", err)
			render.R.JSON(w, 400, map[string]interface{}{"error": "The user ID could not be parsed from the request URL."})
			return
		}
		user.ID = int64(id)
		user, err := user.Get()
		if err != nil {
			log.Println("route ReadUser, user.Get:", err)
			if err.Error() == "not found" {
				render.R.JSON(w, 404, map[string]interface{}{"error": "Not found"})
				return
			}
			render.R.JSON(w, 500, map[string]interface{}{"error": "Internal server error"})
			return
		}
		if len(user.Posts) == 0 {
			p := make([]Post, 0)
			user.Posts = p
		}
		render.R.JSON(w, 200, user)
		return
	case "user":
		var user User
		id, ok := SessionGetValue(r, "id")
		if !ok {
			log.Println("route ReadUser, SessionGetValue:", ok)
			SessionDelete(w, r, "id")
			render.R.HTML(w, 500, "error", "Session could not be fetched. Please log in again.")
			return
		}
		user.ID = id
		user, err := user.Get()
		if err != nil {
			log.Println("route ReadUser, user.Get:", err)
			SessionDelete(w, r, "id")
			render.R.HTML(w, 500, "error", err)
			return
		}
		render.R.HTML(w, 200, "user/index", user)
		return
	}
}

// ReadUsers is a route only available on API side, which fetches all users with post data merged.
// Returns complete list of users on success.
func ReadUsers(w http.ResponseWriter, r *http.Request) {
	var user User
	users, err := user.GetAll()
	if err != nil {
		log.Println("route ReadUsers, user.GetAll:", err)
		render.R.JSON(w, 500, err)
		return
	}
	if len(users) == 0 {
		users = make([]User, 0)
		render.R.JSON(w, 200, users)
		return
	}
	for _, user := range users {
		published := make([]Post, 0)
		if len(user.Posts) == 0 {
			for _, post := range user.Posts {
				if post.Published {
					published = append(published, post)
				}
			}
		}
		user.Posts = published
	}
	render.R.JSON(w, 200, users)
}

// LoginUser is a route which compares plaintext password sent with POST request with
// hash stored in database. On successful request returns session cookie named "user", which contains
// user's ID encrypted, which is the primary key used in database table.
// When called by API it responds with user struct.
// On frontend call it redirects the client to "/user" page.
func LoginUser(w http.ResponseWriter, r *http.Request) {

	user, err := GetUser(r)
	if err != nil {
		log.Println("route LoginUser, user context:", err)
		render.R.JSON(w, 500, map[string]interface{}{"error": "Internal server error"})
		return
	}

	switch Root(r) {
	case "api":
		user, err := user.Login()
		if err != nil {
			log.Println("route LoginUser, user.Login:", err)
			if err.Error() == "wrong username or password" {
				render.R.JSON(w, 401, map[string]interface{}{"error": "Wrong username or password."})
				return
			}
			if err.Error() == "not found" {
				render.R.JSON(w, 404, map[string]interface{}{"error": "User with that email does not exist."})
				return
			}
			render.R.JSON(w, 500, map[string]interface{}{"error": "Internal server error"})
			return
		}

		SessionSetValue(w, r, "id", user.ID)

		user.Password = ""
		render.R.JSON(w, 200, user)
		return
	case "user":
		user, err := user.Login()
		if err != nil {
			log.Println("route LoginUser, user.Login:", err)
			if err.Error() == "wrong username or password" {
				render.R.HTML(w, 401, "user/login", "Wrong username or password.")
				return
			}
			if err.Error() == "not found" {
				render.R.HTML(w, 404, "user/login", "User with that email does not exist.")
				return
			}
			render.R.HTML(w, 500, "user/login", "Internal server error. Please try again.")
			return
		}

		SessionSetValue(w, r, "id", user.ID)

		http.Redirect(w, r, "/user", 302)
		return
	}
}

// RecoverUser is a route of the first step of account recovery, which sends out the recovery
// email etc. associated function calls.
func RecoverUser(w http.ResponseWriter, r *http.Request) {

	user, err := GetUser(r)
	if err != nil {
		log.Println("route CreateUser, user context:", err)
		render.R.JSON(w, 500, map[string]interface{}{"error": "Internal server error"})
		return
	}

	err = user.Recover()
	if err != nil {
		log.Println("route RecoverUser, user.Recover:", err)
		if err.Error() == "not found" {
			render.R.JSON(w, 401, map[string]interface{}{"error": "User with that email does not exist."})
			return
		}
		render.R.JSON(w, 500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	switch Root(r) {
	case "api":
		render.R.JSON(w, 200, map[string]interface{}{"success": "We've sent you a link to your email which you may use you reset your password."})
		return
	case "user":
		http.Redirect(w, r, "/user/login", 302)
		return
	}
}

// ResetUserPassword is a route which is called when accessing the page generated dispatched with
// account recovery emails.
func ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(vestigo.Param(r, "id"))
	if err != nil {
		log.Println("route ResetUserPassword, strconv.Atoi:", err)
		render.R.JSON(w, 400, map[string]interface{}{"error": "User ID could not be parsed from request URL."})
		return
	}

	var user User
	user.ID = int64(id)

	entry, err := user.Get()
	if err != nil {
		log.Println("route ResetUserPassword, user.Get:", err)
		if err.Error() == "not found" {
			render.R.JSON(w, 400, map[string]interface{}{"error": "User with that ID does not exist."})
			return
		}
		render.R.JSON(w, 500, map[string]interface{}{"error": "Internal server error"})
		return
	}

	// this ensures that accounts won't be compromised by posting recovery string as empty,
	// which would otherwise result in succesful password reset
	UUID := uuid.Parse(vestigo.Param(r, "recovery"))
	if UUID == nil {
		log.Println("route ResetUserPassword, uuid.Parse:", err)
		log.Println("there was a problem trying to verify password reset UUID for", entry.Email)
		render.R.JSON(w, 400, map[string]interface{}{"error": "Could not parse UUID from the request."})
		return
	}
	if entry.Recovery == vestigo.Param(r, "recovery") {
		newpassword := context.Get(r, "newpassword").(string)
		entry.Password = newpassword
		_, err = user.PasswordReset(entry)
		if err != nil {
			log.Println("route ResetUserPassword, user.PasswordReset:", err)
			render.R.JSON(w, 500, map[string]interface{}{"error": "Internal server error"})
			return
		}
		switch Root(r) {
		case "api":
			render.R.JSON(w, 200, map[string]interface{}{"success": "Password was updated successfully."})
			return
		case "user":
			http.Redirect(w, r, "/user/login", 302)
			return
		}
	}
}

// LogoutUser is a route which deletes session cookie "user", from the given client.
// On API call responds with HTTP 200 body and on frontend the client is redirected to homepage "/".
func LogoutUser(w http.ResponseWriter, r *http.Request) {
	SessionDelete(w, r, "id")
	switch Root(r) {
	case "api":
		render.R.JSON(w, 200, map[string]interface{}{"success": "You've been logged out."})
		return
	case "user":
		http.Redirect(w, r, "/", 302)
		return
	}
}
