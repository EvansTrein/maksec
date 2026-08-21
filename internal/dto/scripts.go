package dto

import "maksec/internal/entity"

type ScriptsCreateRequest struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Template string `json:"template"`
}

type ScriptsCreateResponse struct {
	Script entity.Script `json:"script"`
}
