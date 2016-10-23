package models

import (
  "github.com/revel/revel"
)

type Degree struct {
  DegreeId      int       `json:"degreeId"`
  UniversityId  int       `json:"universityId"`
  Name          string    `json:"name"`
  TotalCredits  int       `json:"totalCredits"`

  // Transient
  University *University  `json:"university"`
  Courses []*Course       `json:"courses"`
}

func (degree *Degree) Validate(v *revel.Validation) {
  v.Required(degree.University)
  v.Required(degree.Name)
  v.Required(degree.TotalCredits)
}
