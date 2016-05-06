// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package builtin

import (
	"bytes"

	"github.com/ubuntu-core/snappy/interfaces"
)

var locationPermanentSlotAppArmor = []byte(`
# Description: Allow operating as the location service. Reserved because this
#  gives privileged access to the system.
# Usage: reserved

# DBus accesses
#include <abstractions/dbus-strict>
dbus (send)
    bus=system
    path=/org/freedesktop/DBus
    interface=org.freedesktop.DBus
    member={Request,Release}Name
    peer=(name=org.freedesktop.DBus, label=unconfined),

dbus (send)
    bus=system
    path=/org/freedesktop/DBus
    interface=org.freedesktop.DBus
    member=GetConnectionUnix{ProcessID,User}
    peer=(label=unconfined),

# Allow binding the service to the requested connection name
dbus (bind)
    bus=system
    name="com.ubuntu.location.Service",

# Allow traffic to/from org.freedesktop.DBus for location service
dbus (receive, send)
    bus=system
    path=/
    interface=org.freedesktop.DBus*,

dbus (receive, send)
    bus=system
    path=/com/ubuntu/location/Service{,/**}
    interface=org.freedesktop.DBus**
    peer=(label=unconfined),
`)

var locationConnectedSlotAppArmor = []byte(`
# Allow connected clients to interact with the service

# Allow the service to host sessions
dbus (bind)
    bus=system
    name="com.ubuntu.location.Service.Session",

# Allow clients to create a session
dbus (receive)
    bus=system
    path=/com/ubuntu/location/Service
    interface=com.ubuntu.location.Service
    member=CreateSessionForCriteria,

# Allow clients to query service properties
dbus (receive)
    bus=system
    path=/com/ubuntu/location/Service
    interface=org.freedesktop.DBus.Properties
    member=Get,

# Allow clients to set service properties
dbus (receive)
    bus=system
    path=/com/ubuntu/location/Service
    interface=org.freedesktop.DBus.Properties
    member=Set,

# Allow clients to request starting/stopping updates
dbus (receive)
    bus=system
    path=/sessions/*
    interface=com.ubuntu.location.Service.Session
    member={Start,Stop}PositionUpdates,

dbus (receive)
    bus=system
    path=/sessions/*
    interface=com.ubuntu.location.Service.Session
    member={Start,Stop}HeadingUpdates,

dbus (receive)
    bus=system
    path=/sessions/*
    interface=com.ubuntu.location.Service.Session
    member={Start,Stop}VelocityUpdates,

# Allow the service to send updates to clients
dbus (send)
    bus=system
    path=/sessions/*
    interface=com.ubuntu.location.Service.Session
    member=UpdatePosition,

dbus (send)
    bus=system
    path=/sessions/*
    interface=com.ubuntu.location.Service.Session
    member=UpdateHeading,

dbus (send)
    bus=system
    path=/sessions/*
    interface=com.ubuntu.location.Service.Session
    member=UpdateVelocity,

dbus (send)
    bus=system
    path=/com/ubuntu/location/Service
    interface=org.freedesktop.DBus.Properties
    member=PropertiesChanged,
`)

var locationConnectedPlugAppArmor = []byte(`
# Description: Allow using location service. Reserved because this gives
#  privileged access to the service.
# Usage: reserved

#include <abstractions/dbus-strict>

# Allow clients to query service properties
dbus (send)
    bus=system
    path=/com/ubuntu/location/Service
    interface=org.freedesktop.DBus.Properties
    member=Get
    peer=(label=###SLOT_SECURITY_TAGS###),

# Allow clients to set service properties
dbus (send)
    bus=system
    path=/com/ubuntu/location/Service
    interface=org.freedesktop.DBus.Properties
    member=Set
    peer=(label=###SLOT_SECURITY_TAGS###),

# Allow clients to create a session
dbus (send)
    bus=system
    path=/com/ubuntu/location/Service
    interface=com.ubuntu.location.Service
    member=CreateSessionForCriteria
    peer=(label=###SLOT_SECURITY_TAGS###),

# Allow clients to request starting/stopping updates
dbus (send)
    bus=system
    path=/sessions/*
    interface=com.ubuntu.location.Service.Session
    member={Start,Stop}PositionUpdates
    peer=(label=###SLOT_SECURITY_TAGS###),

dbus (send)
    bus=system
    path=/sessions/*
    interface=com.ubuntu.location.Service.Session
    member={Start,Stop}HeadingUpdates
    peer=(label=###SLOT_SECURITY_TAGS###),

dbus (send)
    bus=system
    path=/sessions/*
    interface=com.ubuntu.location.Service.Session
    member={Start,Stop}VelocityUpdates
    peer=(label=###SLOT_SECURITY_TAGS###),

# Allow clients to receive updates from the service
dbus (receive)
    bus=system
    path=/sessions/*
    interface=com.ubuntu.location.Service.Session
    member=UpdatePosition
    peer=(label=###SLOT_SECURITY_TAGS###),

dbus (receive)
    bus=system
    path=/sessions/*
    interface=com.ubuntu.location.Service.Session
    member=UpdateHeading
    peer=(label=###SLOT_SECURITY_TAGS###),

dbus (receive)
    bus=system
    path=/sessions/*
    interface=com.ubuntu.location.Service.Session
    member=UpdateVelocity
    peer=(label=###SLOT_SECURITY_TAGS###),

dbus (receive)
   bus=system
   path=/com/ubuntu/location/Service
   interface=org.freedesktop.DBus.Properties
   member=PropertiesChanged
   peer=(label=###SLOT_SECURITY_TAGS###),

dbus (receive)
    bus=system
    path=/
    interface=org.freedesktop.DBus.ObjectManager
    peer=(label=unconfined),
`)

var locationPermanentSlotSecComp = []byte(`
getsockname
recvmsg
sendmsg
sendto
`)

var locationConnectedPlugSecComp = []byte(`
getsockname
recvmsg
sendmsg
sendto
`)

var locationPermanentSlotDBus = []byte(`
<policy user="root">
    <allow own="com.ubuntu.location.Service"/>
    <allow own="com.ubuntu.location.Service.Session"/>
    <allow send_destination="com.ubuntu.location.Service"/>
    <allow send_destination="com.ubuntu.location.Service.Session"/>
    <allow send_interface="com.ubuntu.location.Service"/>
    <allow send_interface="com.ubuntu.location.Service.Session"/>
</policy>
`)

var locationConnectedPlugDBus = []byte(`
<policy context="default">
    <deny own="com.ubuntu.location.Service"/>
    <allow send_destination="com.ubuntu.location.Service"/>
    <allow send_destination="com.ubuntu.location.Service.Session"/>
    <allow send_interface="com.ubuntu.location.Service"/>
    <allow send_interface="com.ubuntu.location.Service.Session"/>
</policy>
`)

type LocationInterface struct{}

func (iface *LocationInterface) Name() string {
	return "location"
}

func (iface *LocationInterface) PermanentPlugSnippet(plug *interfaces.Plug, securitySystem interfaces.SecuritySystem) ([]byte, error) {
	switch securitySystem {
	case interfaces.SecurityDBus, interfaces.SecurityAppArmor, interfaces.SecuritySecComp, interfaces.SecurityUDev:
		return nil, nil
	default:
		return nil, interfaces.ErrUnknownSecurity
	}
}

func (iface *LocationInterface) ConnectedPlugSnippet(plug *interfaces.Plug, slot *interfaces.Slot, securitySystem interfaces.SecuritySystem) ([]byte, error) {
	switch securitySystem {
	case interfaces.SecurityAppArmor:
		old := []byte("###SLOT_SECURITY_TAGS###")
		new := slotAppLabelExpr(slot)
		snippet := bytes.Replace(locationConnectedPlugAppArmor, old, new, -1)
		return snippet, nil
	case interfaces.SecurityDBus:
		return locationConnectedPlugDBus, nil
	case interfaces.SecuritySecComp:
		return locationConnectedPlugSecComp, nil
	case interfaces.SecurityUDev:
		return nil, nil
	default:
		return nil, interfaces.ErrUnknownSecurity
	}
}

func (iface *LocationInterface) PermanentSlotSnippet(slot *interfaces.Slot, securitySystem interfaces.SecuritySystem) ([]byte, error) {
	switch securitySystem {
	case interfaces.SecurityAppArmor:
		return locationPermanentSlotAppArmor, nil
	case interfaces.SecurityDBus:
		return locationPermanentSlotDBus, nil
	case interfaces.SecuritySecComp:
		return locationPermanentSlotSecComp, nil
	case interfaces.SecurityUDev:
		return nil, nil
	default:
		return nil, interfaces.ErrUnknownSecurity
	}
}

func (iface *LocationInterface) ConnectedSlotSnippet(plug *interfaces.Plug, slot *interfaces.Slot, securitySystem interfaces.SecuritySystem) ([]byte, error) {
	switch securitySystem {
	case interfaces.SecurityAppArmor:
		return locationConnectedSlotAppArmor, nil
	case interfaces.SecurityDBus, interfaces.SecuritySecComp, interfaces.SecurityUDev:
		return nil, nil
	default:
		return nil, interfaces.ErrUnknownSecurity
	}
}

func (iface *LocationInterface) SanitizePlug(slot *interfaces.Plug) error {
	return nil
}

func (iface *LocationInterface) SanitizeSlot(slot *interfaces.Slot) error {
	return nil
}

func (iface *LocationInterface) AutoConnect() bool {
	return false
}
