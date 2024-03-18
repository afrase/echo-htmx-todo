package main

import (
	"html/template"
	"io"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data any, _ echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.gohtml")),
	}
}

var toDoID = 0

type ToDo struct {
	ID          int
	Description string
	Complete    bool
}

func newToDo(description string, complete bool) ToDo {
	toDoID++
	return ToDo{
		ID:          toDoID,
		Description: description,
		Complete:    complete,
	}
}

type ToDos = []ToDo

type Data struct {
	ToDos ToDos
}

func (d *Data) hasToDo(description string) bool {
	for _, toDo := range d.ToDos {
		if toDo.Description == description {
			return true
		}
	}
	return false
}

func (d *Data) addToDo(toDo ToDo) {
	d.ToDos = append(d.ToDos, toDo)
}

func (d *Data) removeToDo(id int) {
	for i, toDo := range d.ToDos {
		if toDo.ID == id {
			d.ToDos = append(d.ToDos[:i], d.ToDos[i+1:]...)
		}
	}
}

func (d *Data) getToDoByID(id int) *ToDo {
	for i, toDo := range d.ToDos {
		if toDo.ID == id {
			return &d.ToDos[i]
		}
	}
	return nil
}

func newData() Data {
	return Data{
		ToDos: ToDos{
			newToDo("Clean room", true),
			newToDo("Pick up groceries", false),
			newToDo("Learn a bit of HTMX", true),
		},
	}
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
	Data     Data
	FormData FormData
}

func newPage() Page {
	return Page{
		Data:     newData(),
		FormData: newFormData(),
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Static("/css", "css")
	e.Static("/img", "img")
	e.Renderer = newTemplate()

	page := newPage()

	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", page)
	})

	e.POST("/todos", func(c echo.Context) error {
		description := c.FormValue("description")

		if description == "" {
			page.FormData.Values["description"] = description
			page.FormData.Errors["description"] = "Description can't be blank"
			return c.Render(422, "createToDo", page.FormData)
		}

		if page.Data.hasToDo(description) {
			page.FormData.Values["description"] = description
			page.FormData.Errors["description"] = "ToDo already exists"
			return c.Render(422, "createToDo", page.FormData)
		}

		toDo := newToDo(description, false)
		page.Data.addToDo(toDo)

		err := c.Render(200, "createToDo", newFormData())
		if err != nil {
			return err
		}
		return c.Render(200, "oobToDo", toDo)
	})

	e.GET("/todos/:id/toggle", func(c echo.Context) error {
		strID := c.Param("id")
		ID, err := strconv.Atoi(strID)
		if err != nil {
			return c.String(400, "Invalid ID")
		}
		toDo := page.Data.getToDoByID(ID)
		if toDo == nil {
			return c.String(404, "ToDo not found")
		}
		toDo.Complete = !toDo.Complete
		return c.Render(200, "toDo", toDo)
	})

	e.DELETE("/todos/:id", func(c echo.Context) error {
		strID := c.Param("id")
		ID, err := strconv.Atoi(strID)
		if err != nil {
			return c.String(400, "Invalid ID")
		}
		page.Data.removeToDo(ID)
		return c.NoContent(200)
	})

	e.Logger.Fatal(e.Start(":8080"))
}
