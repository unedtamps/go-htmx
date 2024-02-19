package main

import (
	"html/template"
	"io"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("view/*.html")),
	}
}

type Contact struct {
	Id    int
	Name  string
	Email string
}

var id = 0

func newContact(name, email string) Contact {
	id++
	return Contact{
		Id:    id,
		Name:  name,
		Email: email,
	}
}

type Contacts = []Contact

type Data struct {
	Contacts Contacts
}

func newData() Data {
	return Data{
		Contacts: []Contact{
			newContact("Gopher", "json@gmail.com"),
			newContact("Unedo", "unedo@gmail.com"),
		},
	}
}

func (d *Data) indexOfData(id int) int {
	for i, c := range d.Contacts {
		if c.Id == id {
			return i
		}
	}
	return -1
}

func (d *Data) hasEmail(email string) bool {
	for _, c := range d.Contacts {
		if c.Email == email {
			return true
		}
	}
	return false
}

type FormData struct {
	Values map[string]string
	Errors map[string]string
}

func newFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

type Page struct {
	Data Data
	Form FormData
}

func newPage() Page {
	return Page{
		Data: newData(),
		Form: newFormData(),
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Static("/", "public")

	e.Renderer = newTemplate()

	page := newPage()

	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", page)
	})

	e.POST("/contacts", func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")

		if page.Data.hasEmail(email) == true {
			formData := newFormData()
			formData.Errors["email"] = "Email already exists"
			formData.Values["name"] = name
			formData.Values["email"] = email
			return c.Render(422, "form", formData)
		}
		contacts := newContact(name, email)
		page.Data.Contacts = append(page.Data.Contacts, contacts)
		c.Render(200, "form", newFormData())
		return c.Render(200, "oob-list", contacts)
	})

	e.DELETE("/contacts/:id", func(c echo.Context) error {
		time.Sleep(5 * time.Second)
		idstr := c.Param("id")
		id, err := strconv.Atoi(idstr)
		if err != nil {
			return err
		}
		index := page.Data.indexOfData(id)
		if index == -1 {
			return c.String(404, "Not Found")
		}
		page.Data.Contacts = append(page.Data.Contacts[:index], page.Data.Contacts[index+1:]...)
		return c.NoContent(200)
	})

	e.Logger.Fatal(e.Start(":8080"))
}
