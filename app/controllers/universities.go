package controllers

import (
  "github.com/revel/revel"
  "credREST/app/models"
  "encoding/json"
  "net/http"
  jwt "github.com/dgrijalva/jwt-go"
  "fmt"
  "errors"
)

type Universities struct {
  Application
}

func (c Universities) Index() revel.Result {
  results, err := c.Txn.Select(models.University{},
    `select * from "University"`)
  if err != nil {
    panic(err)
  }

  var universities []*models.University
  for _, r := range results {
    u := r.(*models.University)
    universities = append(universities, u)
  }

  return c.RenderJson(universities)
}

func (c Universities) loadUniversityById(id int) *models.University {
  u, err := c.Txn.Get(models.University{}, id)
  if err != nil {
    panic(err)
  }
  if u == nil {
    return nil
  }
  return u.(*models.University)
}

func (c Universities) loadUniversityWithDegrees(id int) *models.University {
  u := c.loadUniversityById(id)
  if u == nil {
    return nil
  }

  // Get all the university's degrees.
  var degrees []*models.Degree
  _, err := c.Txn.Select(&degrees,
    `select * from "Degree" where "UniversityId"=$1;`, id)
  if err != nil {
    panic(err)
  }
  u.Degrees = append(u.Degrees, degrees...)

  return u
}

func (c Universities) Show(id int) revel.Result {
  university := c.loadUniversityWithDegrees(id)
  if university == nil {
    return c.NotFound("University %d does not exist", id)
  }
  return c.RenderJson(university)
}

func (c Universities) parseUniversity() (models.University, error) {
  university := models.University{}
  err := json.NewDecoder(c.Request.Body).Decode(&university)
  return university, err
}

func (c Universities) Create() revel.Result {

  // Check if the request is authorized by a JSON Web Token.
  tokenString := c.Request.Header.Get("Authorization")

  if err := c.checkJWTUniversity(tokenString); err != nil {
    c.Response.Status = http.StatusUnauthorized
    c.Response.ContentType = "application/json"
    // Change it to send a JSON with the message of the error, use the error 'err'.
    return c.RenderText("ERROR: You dont have authorization to perform this action.")
  }

  if university, err := c.parseUniversity(); err != nil {
    return c.RenderText("Unable to parse a University from JSON.")
  } else {
    // Validate the model
    university.Validate(c.Validation)
    if c.Validation.HasErrors() {
      // Do something better here!
      return c.RenderText("You have an error in your University.")
    } else {
      if err := c.Txn.Insert(&university); err != nil {
        return c.RenderText("Error inserting record.")
      } else {
        c.Response.Status = http.StatusCreated
        c.Response.ContentType = "application/json"
        return c.RenderJson(university)
      }
    }
  }
}

func (c Universities) Update(id int) revel.Result {
  // Check if the request is authorized by a JSON Web Token.
  tokenString := c.Request.Header.Get("Authorization")

  if err := c.checkJWTUniversity(tokenString); err != nil {
    c.Response.Status = http.StatusUnauthorized
    c.Response.ContentType = "application/json"
    // Change it to send a JSON with the message of the error, use the error 'err'.
    return c.RenderText("ERROR: You dont have authorization to perform this action.")
  }

  university, err := c.parseUniversity()
  if err != nil {
    return c.RenderText("Unable to parse a University from the JSON.")
  }
  // Ensure the Id is set.
  university.UniversityId = id
  success, err := c.Txn.Update(&university)
  if err != nil || success == 0 {
    return c.RenderText("Unable to update the University")
  }
  return c.RenderJson(university)
}

func (c Universities) Delete(id int) revel.Result {
  // Check if the request is authorized by a JSON Web Token.
  tokenString := c.Request.Header.Get("Authorization")

  if err := c.checkJWTUniversity(tokenString); err != nil {
    c.Response.Status = http.StatusUnauthorized
    c.Response.ContentType = "application/json"
    // Change it to send a JSON with the message of the error, use the error 'err'.
    return c.RenderText("ERROR: You dont have authorization to perform this action.")
  }

  success, err := c.Txn.Delete(&models.University{ UniversityId: id })
  if err != nil || success == 0 {
    return c.RenderText("Failed to remove the University.")
  }
  msg := "{ message: 'Eliminado.'}"
  return c.RenderJson(msg)
}

func (c Universities) checkJWTUniversity(tokenString string) error {
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
