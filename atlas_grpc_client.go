package main


func (atlas *Atlas) getESP1(stop chan any, seq int64) chan *ESP1PharmNDC {
	return strm_recv_srvr[ESP1PharmNDC]("atlas", "esp1", seq, atlas.X509cert, atlas.titan.GetESP1Pharms, stop);
}
func (atlas *Atlas) getEntities(stop chan any, seq int64) chan *Entity {
	return strm_recv_srvr[Entity]("atlas", "ents", seq, atlas.X509cert, atlas.titan.GetEntities, stop);
}
func (atlas *Atlas) getLedger(stop chan any, seq int64) chan *Eligibility {
	return strm_recv_srvr[Eligibility]("atlas", "elig", seq, atlas.X509cert, atlas.titan.GetEligibilityLedger, stop);
}
func (atlas *Atlas) getNDCs(stop chan any, seq int64) chan *NDC {
	return strm_recv_srvr[NDC]("atlas", "ndcs", seq, atlas.X509cert, atlas.titan.GetNDCs, stop);
}
func (atlas *Atlas) getPharms(stop chan any, seq int64) chan *Pharmacy {
	return strm_recv_srvr[Pharmacy]("atlas", "phms", seq, atlas.X509cert, atlas.titan.GetPharmacies, stop);
}
func (atlas *Atlas) getSPIs(stop chan any, seq int64) chan *SPI {
	return strm_recv_srvr[SPI]("atlas", "spis", seq, atlas.X509cert, atlas.titan.GetSPIs, stop);
}
func (atlas *Atlas) getDesignations(stop chan any, seq int64) chan *Designation {
	return strm_recv_srvr[Designation]("atlas", "desigs", seq, atlas.X509cert, atlas.titan.GetDesignations, stop);
}
func (atlas *Atlas) getLDNs(stop chan any, seq int64) chan *LDN {
	return strm_recv_srvr[LDN]("atlas", "ldns", seq, atlas.X509cert, atlas.titan.GetLDNs, stop);
}
