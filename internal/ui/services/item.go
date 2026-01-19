package services

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"github.com/givensuman/containertui/internal/client"
)

type ServiceItem struct {
	Service client.Service
}

func (i ServiceItem) Title() string {
	return i.Service.Name
}

func (i ServiceItem) Description() string {
	return fmt.Sprintf("Replicas: %d | Containers: %d", i.Service.Replicas, len(i.Service.Containers))
}

func (i ServiceItem) FilterValue() string {
	return i.Service.Name
}

var (
	_ list.Item        = (*ServiceItem)(nil)
	_ list.DefaultItem = (*ServiceItem)(nil)
)
