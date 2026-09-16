package handlers

import (
	service "fairchild_be/internal/services/contact"
)

type ContactHandler struct {
	contactService *service.ContactService
}

func NewHandler(contactService *service.ContactService) *ContactHandler {
	return &ContactHandler{contactService: contactService}
}
