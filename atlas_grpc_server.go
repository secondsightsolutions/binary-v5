package main

import (
	context "context"
	"fmt"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

type atlasServer struct {
	UnimplementedAtlasServer
}

func (s *atlasServer) Ping(ctx context.Context, in *Req) (*Res, error) {
	if err := s.validate(ctx); err != nil {
		return &Res{}, err
	}
	return &Res{}, nil
}

func (s *atlasServer) NewScrub(ctx context.Context, req *Scrub) (*ScrubRes, error) {
	if err := s.validate(ctx); err != nil {
		return &ScrubRes{}, err
	}
	if scid, err := db_insert_one(ctx, atlas.pools["atlas"], "atlas.scrubs", nil, req, "scid"); err == nil {
		return &ScrubRes{Scid: scid}, nil
	} else {
		return nil, err
	}
}

func (s *atlasServer) Rebates(strm grpc.ClientStreamingServer[Rebate, Res]) error {
	for {
		if rbt, err := strm.Recv(); err == nil {
			fmt.Printf("atlas: %v\n", rbt)
		} else {
			break
		}
	}
	return nil
}

func (s *atlasServer) validate(ctx context.Context) error {
	if p, ok := peer.FromContext(ctx); ok && p != nil {
		if tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo); ok {
			cn, ou := getCreds(tlsInfo)
			name, auth, vers, _mnu, _ := getMetaGRPC(ctx)

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
			if cn != name {
				return fmt.Errorf("name on cert doesnt match name in metadata")
			}
			if _mnu != manu {
				return fmt.Errorf("manu in metadata is incorrect")
			}
			qry := `
				FROM atlas.proc p
				JOIN atlas.proc_auth pa ON (p.prid  = pa.prid)
				JOIN atlas.auth a       ON (pa.auth = a.auth)
				WHERE p.enabled  = TRUE
				AND   pa.enabled = TRUE
				AND   a.enabled  = TRUE
				AND   p.prid = '%s'
				AND   a.auth = '%s'
			`
			if cnt, err := db_count(context.Background(), atlas.pools["atlas"], fmt.Sprintf(qry, name, auth)); err == nil {
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
