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

type Degrees struct {
  Application
}

func (c Degrees) Index() revel.Result {
  results, err := c.Txn.Select(models.Degree{},
    `select * from "Degree"`)
  if err != nil {
    panic(err)
  }

  var degrees []*models.Degree
  for _, r := range results {
    d := r.(*models.Degree)
    degrees = append(degrees, d)
  }

  return c.RenderJson(degrees)
}

func (c Degrees) loadDegreeById(id int) *models.Degree {
  d, err := c.Txn.Get(models.Degree{}, id)
  if err != nil {
    panic(err)
  }
  if d == nil {
    return nil
  }

  u, err := c.Txn.Get(models.University{}, d.(*models.Degree).UniversityId)
  if err != nil {
    panic(err)
  }
  if u == nil {
    return nil
  }
  d.(*models.Degree).University = u.(*models.University)
  return d.(*models.Degree)
}

func (c Degrees) loadDegreeWithCourses(id int) *models.Degree {
  d := c.loadDegreeById(id)
  if d == nil {
    return nil
  }

  // Get all the degree's courses.
  var courses []*models.Course
  _, err := c.Txn.Select(&courses,
    `select * from "Course" where "DegreeId"=$1;`, id)
  if err != nil {
    panic(err)
  }
  d.Courses = append(d.Courses, courses...)

  return d
}

func (c Degrees) Show(id int) revel.Result {
  degree := c.loadDegreeWithCourses(id)
  if degree == nil {
    return c.NotFound("Degree %d does not exist", id)
  }
  return c.RenderJson(degree)
}


func (c Degrees) parseDegree() (models.Degree, error) {
  degree := models.Degree{}
  err := json.NewDecoder(c.Request.Body).Decode(&degree)
  return degree, err
}

func (c Degrees) Create() revel.Result {
  // Check if the request is authorized by a JSON Web Token.
  tokenString := c.Request.Header.Get("Authorization")

  if err := c.checkJWTDegrees(tokenString); err != nil {
    c.Response.Status = http.StatusUnauthorized
    c.Response.ContentType = "application/json"
    // Change it to send a JSON with the message of the error, use the error 'err'.
    return c.RenderText("ERROR: You dont have authorization to perform this action.")
  }

  if degree, err := c.parseDegree(); err != nil {
    return c.RenderText("Unable to parse a degree from the JSON.")
  } else {
    // Validate the model.
    degree.Validate(c.Validation)
    if c.Validation.HasErrors() {
      // Do something better here!
      return c.RenderText("You have an error in your degree.")
    } else {
      if err := c.Txn.Insert(&degree); err != nil {
        return c.RenderText("Error inserting record for the degree.")
      } else {
        c.Response.Status = http.StatusCreated
        c.Response.ContentType = "application/json"
        return c.RenderJson(degree)
      }
    }
  }
}

func (c Degrees) Update(id int) revel.Result {
  // Check if the request is authorized by a JSON Web Token.
  tokenString := c.Request.Header.Get("Authorization")

  if err := c.checkJWTDegrees(tokenString); err != nil {
    c.Response.Status = http.StatusUnauthorized
    c.Response.ContentType = "application/json"
    // Change it to send a JSON with the message of the error, use the error 'err'.
    return c.RenderText("ERROR: You dont have authorization to perform this action.")
  }

  degree, err := c.parseDegree()
  if err != nil {
    return c.RenderText("Unable to parse a Degree from the JSON.")
  }
  // Ensure the Id is set.
  degree.DegreeId = id
  success, err := c.Txn.Update(&degree)
  if err != nil || success == 0 {
    return c.RenderText("Unable to update the Degree.")
  }
  return c.RenderJson(degree)
}

func (c Degrees) Delete(id int) revel.Result {
  // Check if the request is authorized by a JSON Web Token.
  tokenString := c.Request.Header.Get("Authorization")

  if err := c.checkJWTDegrees(tokenString); err != nil {
    c.Response.Status = http.StatusUnauthorized
    c.Response.ContentType = "application/json"
    // Change it to send a JSON with the message of the error, use the error 'err'.
    return c.RenderText("ERROR: You dont have authorization to perform this action.")
  }

  success, err := c.Txn.Delete(&models.Degree{ DegreeId: id })
  if err != nil || success == 0 {
    return c.RenderText("Failed to remove the Degree.")
  }
  msg := "{ message: 'Eliminado.'}"
  return c.RenderJson(msg)
}

func (c Degrees) checkJWTDegrees(tokenString string) error {
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
