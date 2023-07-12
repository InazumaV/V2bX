package format

import (
	"fmt"
)

func UserTag(tag string, uuid string) string {
	return fmt.Sprintf("%s|%s", tag, uuid)
}
