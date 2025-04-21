package http_profile

import (
	"avito/internal/domain"
	"log"
)

func receptionStatus(status domain.ReceptionStatus) ReceptionStatus {
	switch status {
	case domain.CloseProductAcceptanceStatus:
		return "close"
	case domain.InProggressProductAcceptanceStatus:
		return "in_progress"
	default:
		log.Println("fall of possible ReceptionStatuses")
		return ""
	}
}

func productType(category domain.ProductCategory) ProductType {
	switch category {
	case domain.ElectronicsProductCategory:
		return "электроника"
	case domain.ClothesProductCategory:
		return "одежда"
	case domain.ShoesProductCategory:
		return "обувь"
	default:
		log.Println("fall out of known product categories")
		return ""
	}
}

func role(roleId domain.UserRoleID) UserRole {
	switch roleId {
	case domain.ClientUserRoleID:
		return UserRoleEmployee
	case domain.ModeratorUserRoleID:
		return UserRoleModerator
	default:
		log.Println("fall out of known user roles")
		return ""
	}
}
