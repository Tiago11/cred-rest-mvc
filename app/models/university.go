package models

import (
  "github.com/revel/revel"
)

type University struct {
  UniversityId    int       `json:"universityId"`
  Name            string    `json:"name"`
  Country         string    `json:"country"`

  // Transient
  Degrees  []*Degree         `json:"degrees"`
}

func (university University) Validate(v *revel.Validation) {
  v.Required(university.Name)
  v.Required(university.Country)
}
