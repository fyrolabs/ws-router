package ws

import "github.com/olahol/melody"

// Context is the context for a request.
// use c.Get(key) and c.Set(key) to get and set session data
// c.Respond(data) to send a response
type Context struct {
	Request *Request
	session *melody.Session
}

func (c *Context) Get(key string) any {
	value, exists := c.session.Get(key)
	if !exists {
		return nil
	}
	return value
}

// GetString returns the value as a string, or an empty string if the value is nil or not a string
func (c *Context) GetString(key string) string {
	value := c.Get(key)
	if value == nil {
		return ""
	}

	strValue, ok := value.(string)
	if !ok {
		return ""
	}

	return strValue
}

// GetInt returns the value as an int, or 0 if the value is nil or not an int
func (c *Context) GetInt(key string) int {
	value := c.Get(key)
	if value == nil {
		return 0
	}

	intValue, ok := value.(int)
	if !ok {
		return 0
	}

	return intValue
}

func (c *Context) Set(key string, value any) {
	c.session.Set(key, value)
}
