package bridge

import (
	"fmt"
)

func (a AuthContext) Validate() error {
	if a.Acting != nil && a.Base == nil {
		return fmt.Errorf("acting identity requires a base profile")
	}
	if a.Base != nil && a.Base.ID == "" {
		return fmt.Errorf("base profile ID is required")
	}
	if a.Console != nil && a.Console.ID == "" {
		return fmt.Errorf("console profile ID is required")
	}
	return nil
}

func (a AuthContext) EffectiveEmail() string {
	if a.Acting != nil {
		return a.Acting.Email
	}
	if a.Base != nil {
		return a.Base.Email
	}
	return ""
}
