package models

import (
  "github.com/revel/revel"
)

type Course struct {
  CourseId      int       `json:"courseId"`
  DegreeId      int       `json:"degreeId"`
  Name          string    `json:"name"`
  Credits       int       `json:"credits"`

  // Transient
  Degree  *Degree         `json:"degree"`
}

func (course *Course) Validate(v *revel.Validation){
  v.Required(course.Degree)
  v.Required(course.Name)
  v.Required(course.Credits)
}
