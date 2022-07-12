package notepet

type AuthFunc func(user, pass string) (bool, error)
