package problem

import (
	"errors"
	"hash/fnv"
	"strconv"
	"strings"
)

const defaultRouteSalt = "luooj-problem-route"

type RouteCodec struct {
	mask uint64
}

func NewRouteCodec(salt string) RouteCodec {
	normalizedSalt := strings.TrimSpace(salt)
	if normalizedSalt == "" {
		normalizedSalt = defaultRouteSalt
	}

	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(normalizedSalt))

	return RouteCodec{mask: hasher.Sum64()}
}

func (c RouteCodec) Encode(problemNo int64) string {
	if problemNo <= 0 {
		return ""
	}
	return strings.ToUpper(strconv.FormatUint(uint64(problemNo)^c.mask, 36))
}

func (c RouteCodec) Decode(routeCode string) (int64, error) {
	normalizedCode := strings.TrimSpace(routeCode)
	if normalizedCode == "" {
		return 0, errors.New("route code is required")
	}

	decoded, err := strconv.ParseUint(strings.ToLower(normalizedCode), 36, 64)
	if err != nil {
		return 0, errors.New("invalid route code")
	}

	problemNo := int64(decoded ^ c.mask)
	if problemNo <= 0 {
		return 0, errors.New("invalid route code")
	}

	return problemNo, nil
}
