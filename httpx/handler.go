package httpx

import "net/http"

type Req = http.Request
type Res = http.ResponseWriter

type Handler func(Res, *Req) error
