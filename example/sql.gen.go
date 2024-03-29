// Code generated by go generate; DO NOT EDIT.
package sql

type AuthorsCreate struct {
	Bio  interface{}
	Name interface{}
}

type AuthorsFetch struct {
	AuthorId interface{}
}

type AuthorsList struct {
}

type BooksFetch struct {
	BookId interface{}
}

type BooksList struct {
}

func (p AuthorsCreate) Build() (string, []interface{}) {
	return "INSERT INTO authors (\n  name, bio\n) VALUES (\n  $1, $2\n)\nRETURNING *;", []interface{}{
		p.Name,
		p.Bio,
	}
}

func (p AuthorsFetch) Build() (string, []interface{}) {
	return "SELECT * FROM authors\nWHERE id = $1 LIMIT 1;", []interface{}{
		p.AuthorId,
	}
}

func (p AuthorsList) Build() (string, []interface{}) {
	return "SELECT * FROM authors\nORDER BY name;", []interface{}{}
}

func (p BooksFetch) Build() (string, []interface{}) {
	return "SELECT * FROM book\nWHERE book_id = $1;", []interface{}{
		p.BookId,
	}
}

func (p BooksList) Build() (string, []interface{}) {
	return "", []interface{}{}
}
