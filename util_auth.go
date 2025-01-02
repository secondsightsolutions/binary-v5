package main

import (
	context "context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func validate_client(ctx context.Context, pool *pgxpool.Pool, schm string) error {
	if p, ok := peer.FromContext(ctx); ok && p != nil {
		if tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo); ok {
			cn, ou := getCreds(tlsInfo)
			name, auth, vers, _mnu, kind, _ := getMetaGRPC(ctx)

			if cn == "" {
				return fmt.Errorf("missing CN from cert (or no cert presented)")
			}
			if ou == "" {
				return fmt.Errorf("missing OU from cert (or no cert presented)")
			}
			if name == "" {
				return fmt.Errorf("missing name from metadata")
			}
			if auth == "" {
				return fmt.Errorf("missing auth from metadata")
			}
			if vers == "" {
				return fmt.Errorf("missing vers from metadata")
			}
			if _mnu == "" {
				return fmt.Errorf("missing manu from metadata")
			}
			if kind == "" {
				return fmt.Errorf("missing kind from metadata")
			}
			if cn != name {
				return fmt.Errorf("name on cert doesnt match name in metadata")
			}
			if _mnu != manu && name != "brg" {
				return fmt.Errorf("manu in metadata is incorrect")
			}
			qry := `
				FROM  %s.auth
				WHERE enb  = TRUE
				AND   manu = '%s'
				AND   proc = '%s'
				AND   auth = '%s'
				AND   kind = '%s'
			`
			if cnt, err := db_count(context.Background(), pool, fmt.Sprintf(qry, schm, manu, name, auth, kind)); err == nil {
				if cnt == 0 {
					return fmt.Errorf("not authorized")
				}
			} else {
				return err
			}
			return nil
		} else {
			return fmt.Errorf("invalid TLS info")
		}
	} else {
		return fmt.Errorf("missing peer in context")
	}
}

func metaGRPC(args map[string]string) context.Context {
	md  := metadata.Pairs("name", name, "auth", opts.auth, "kind", opts.kind, "vers", vers, "manu", manu)
	for k, v := range args {
		md.Set(k, v)
	}
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	return ctx
}
func getMetaGRPC(ctx context.Context) (name, auth, vers, manu, kind, scid string) {
	val := func(md metadata.MD, name string) string {
		if vals, ok := md[name]; ok {
			if len(vals) > 0 {
				return vals[0]
			}
		}
		return ""
	}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		name = val(md, "name")
		auth = val(md, "auth")
		vers = val(md, "vers")
		manu = val(md, "manu")
		scid = val(md, "scid")
		kind = val(md, "kind")
	}
	return
}
