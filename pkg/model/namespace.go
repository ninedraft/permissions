package model

import (
	"time"

	"git.containerum.net/ch/permissions/pkg/errors"
	"github.com/containerum/kube-client/pkg/model"
	"github.com/go-pg/pg/orm"
	"github.com/satori/go.uuid"
)

// Namespace describes namespace
//
// swagger:model
type Namespace struct {
	tableName struct{} `sql:"namespaces"`

	Resource

	// swagger:strfmt uuid
	TariffID       *string `sql:"tariff_id,type:uuid" json:"tariff_id,omitempty"`
	RAM            int     `sql:"ram,notnull" json:"ram"`
	CPU            int     `sql:"cpu,notnull" json:"cpu"`
	MaxExtServices int     `sql:"max_ext_services,notnull" json:"max_external_services"`
	MaxIntServices int     `sql:"max_int_services,notnull" json:"max_internal_services"`
	MaxTraffic     int     `sql:"max_traffic,notnull" json:"max_traffic"`
	KubeName       string  `sql:"kube_name,unique:kube_name,notnull" json:"kube_name"`
	// swagger:strfmt uuid
	ProjectID *string `sql:"project_id,type:uuid" json:"project_id,omitempty"`
}

func (ns *Namespace) BeforeInsert(db orm.DB) error {
	cnt, err := db.Model(ns).
		Where("owner_user_id = ?owner_user_id").
		Where("label = ?label").
		Where("NOT deleted").
		Count()
	if err != nil {
		return err
	}

	if ns.KubeName == "" {
		ns.KubeName = uuid.NewV4().String()
		ns.ID = ns.KubeName
	}

	if cnt > 0 {
		return errors.ErrResourceAlreadyExists().AddDetailF("namespace %s already exists", ns.Label)
	}

	return nil
}

func (ns *Namespace) AfterInsert(db orm.DB) error {

	return db.Insert(&Permission{
		ResourceID:         ns.ID,
		UserID:             ns.OwnerUserID,
		ResourceType:       ResourceNamespace,
		InitialAccessLevel: model.Owner,
		CurrentAccessLevel: model.Owner,
	})
}

// NamespaceWithPermissions is a response object for get requests
//
// swagger:model NamespaceWithPermissions
type NamespaceWithPermissions struct {
	Namespace `pg:",override"`

	Permission Permission `pg:"fk:resource_id" sql:"-" json:",inline"`

	Permissions []Permission `pg:"polymorphic:resource_" sql:"-" json:"users"`
}

func (np *NamespaceWithPermissions) ToKube() model.Namespace {
	ns := model.Namespace{
		ID:            np.KubeName,
		CreatedAt:     new(string),
		Owner:         np.OwnerUserID,
		OwnerLogin:    np.OwnerUserLogin,
		Label:         np.Label,
		Access:        np.Permission.CurrentAccessLevel,
		MaxExtService: uint(np.MaxExtServices),
		MaxIntService: uint(np.MaxIntServices),
		MaxTraffic:    uint(np.MaxTraffic),
		Resources: model.Resources{
			Hard: model.Resource{
				CPU:    uint(np.CPU),
				Memory: uint(np.RAM),
			},
		},
		Users: make([]model.UserAccess, len(np.Permissions)),
	}
	*ns.CreatedAt = np.CreateTime.Format(time.RFC3339)
	for i, v := range np.Permissions {
		ns.Users[i] = model.UserAccess{
			Username:    v.UserLogin,
			AccessLevel: v.CurrentAccessLevel,
		}
	}
	return ns
}

// NamespaceAdminCreateRequest contains parameters for creating namespace without billing
//
// swagger:model
type NamespaceAdminCreateRequest struct {
	Label          string `json:"label" binding:"required"`
	CPU            int    `json:"cpu" binding:"required"`
	Memory         int    `json:"memory" binding:"required"`
	MaxExtServices int    `json:"max_ext_services" binding:"required"`
	MaxIntServices int    `json:"max_int_services" binding:"required"`
	MaxTraffic     int    `json:"max_traffic" binding:"required"`
}

// NamespaceAdminResizeRequest contains parameter for resizing namespace without billing
//
// swagger:model
type NamespaceAdminResizeRequest struct {
	CPU            *int `json:"cpu"`
	Memory         *int `json:"memory"`
	MaxExtServices *int `json:"max_ext_services"`
	MaxIntServices *int `json:"max_int_services"`
	MaxTraffic     *int `json:"max_traffic"`
}

// NamespaceCreateRequest contains parameters for creating namespace
//
// swagger:model
type NamespaceCreateRequest struct {
	// swagger:strfmt uuid
	TariffID string `json:"tariff_id" binding:"required,uuid"`

	Label string `json:"label" binding:"required"`
}

// NamespaceRenameRequest contains parameters for renaming namespace
//
// swagger:model
type NamespaceRenameRequest = model.ResourceUpdateName

// NamespaceResizeRequest contains parameters for changing namespace quota
//
// swagger:model
type NamespaceResizeRequest struct {
	// swagger:strfmt uuid
	TariffID string `json:"tariff_id" binding:"required,uuid"`
}
