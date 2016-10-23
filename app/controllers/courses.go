package controllers

import (
  "github.com/revel/revel"
  "credREST/app/models"
  "encoding/json"
  "net/http"
  "errors"
  jwt "github.com/dgrijalva/jwt-go"
  "fmt"
)

type Courses struct {
  Application
}

func (c Courses) Index() revel.Result {
  results, err := c.Txn.Select(models.Course{},
    `select * from "Course"`)
  if err != nil {
    panic(err)
  }

  var courses []*models.Course
  for _, r := range results {
    co := r.(*models.Course)
    courses = append(courses, co)
  }

  return c.RenderJson(courses)
}

func (c Courses) loadCourseById(id int) *models.Course {
  co, err := c.Txn.Get(models.Course{}, id)
  if err != nil {
    panic(err)
  }
  if co == nil {
    return nil
  }

  d, err := c.Txn.Get(models.Degree{}, co.(*models.Course).DegreeId)
  if err != nil {
    panic(err)
  }
  if d == nil {
    return nil
  }
  co.(*models.Course).Degree = d.(*models.Degree)

  return co.(*models.Course)
}

func (c Courses) Show(id int) revel.Result {
  course := c.loadCourseById(id)
  if course == nil {
    return c.NotFound("Course %d does not exist", id)
  }
  return c.RenderJson(course)
}

func (c Courses) parseCourse() (models.Course, error) {
  course := models.Course{}
  err := json.NewDecoder(c.Request.Body).Decode(&course)
  return course, err
}

func (c Courses) Create() revel.Result {
  // Check if the request is authorized by a JSON Web Token.
  tokenString := c.Request.Header.Get("Authorization")

  if err := c.checkJWTCourses(tokenString); err != nil {
    c.Response.Status = http.StatusUnauthorized
    c.Response.ContentType = "application/json"
    // Change it to send a JSON with the message of the error, use the error 'err'.
    return c.RenderText("ERROR: You dont have authorization to perform this action.")
  }

  if course, err := c.parseCourse(); err != nil {
    return c.RenderText("Unable to parse the course from JSON.")
  } else {
    // Validate the model.
    course.Validate(c.Validation)
    if c.Validation.HasErrors() {
      // Do something better here!
      return c.RenderText("You have an error in your Course.")
    } else {
      if err := c.Txn.Insert(&course); err != nil {
        return c.RenderText("Error inserting the course.")
      } else {
        c.Response.Status = http.StatusCreated
        c.Response.ContentType = "application/json"
        return c.RenderJson(course)
      }
    }
  }
}

func (c Courses) Update(id int) revel.Result {
  // Check if the request is authorized by a JSON Web Token.
  tokenString := c.Request.Header.Get("Authorization")

  if err := c.checkJWTCourses(tokenString); err != nil {
    c.Response.Status = http.StatusUnauthorized
    c.Response.ContentType = "application/json"
    // Change it to send a JSON with the message of the error, use the error 'err'.
    return c.RenderText("ERROR: You dont have authorization to perform this action.")
  }

  course, err := c.parseCourse()
  if err != nil {
    return c.RenderText("Unable to parse a Course from the JSON.")
  }
  // Ensure the Id is set.
  course.CourseId = id
  success, err := c.Txn.Update(&course)
  if err != nil || success == 0 {
    return c.RenderText("Unable to update the Course.")
  }
  return c.RenderJson(course)
}

func (c Courses) Delete(id int) revel.Result {
  // Check if the request is authorized by a JSON Web Token.
  tokenString := c.Request.Header.Get("Authorization")

  if err := c.checkJWTCourses(tokenString); err != nil {
    c.Response.Status = http.StatusUnauthorized
    c.Response.ContentType = "application/json"
    // Change it to send a JSON with the message of the error, use the error 'err'.
    return c.RenderText("ERROR: You dont have authorization to perform this action.")
  }

  success, err := c.Txn.Delete(&models.Course{ CourseId: id })
  if err != nil || success == 0 {
    return c.RenderText("Failed to remove the Course.")
  }
  msg := "{ message: 'Eliminado.'}"
  return c.RenderJson(msg)
}

func (c Courses) checkJWTCourses(tokenString string) error {
  if tokenString == "" {
    return errors.New("Authorization header not found.")
  }

  token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
    }
    return []byte(revel.Config.StringDefault("jwt.secret", "default")), nil
  })

  if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
    // Esto tiene que ser diferente.
    if claims["admin"] == true {
      return nil
    } else {
      return errors.New("Wrong claim inside JWT.")
    }
  } else {
    return err
  }
}
