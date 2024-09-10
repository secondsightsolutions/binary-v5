package main

var Fields = struct {
    Stat    string  // status
    Auth    string
    Vers    string
    Manu    string  // manufacturer
    Envr    string  // environment - staging, prod, etc.
    Spid    string  // service_provider_id
    Rxn     string  // rx_number (hashed or clear)
    Frxn    string  // formatted_rx_number (hashed or clear) (not really used - moved into Rxn)
    Dos     string  // date_of_service (hashed or clear) - for use by rebate DOS
    Dof     string  // date_of_service (hashed or clear) - fill date, from submission_rows.date_of_service
    Doc     string  // date of claim (created_at)
    Dol     string // rbt_hdos (the hashed rebate DOS set as the ECS lock on the claim)
    Dpr     string  // date_prescribed (hashed or clear)
    Ndc     string  // ndc
    I340    string  // id_340b
    Istc    string  // status_code_340b
    Ceid    string  // contracted_entity_id
    Phid    string  // pharmacy_id
    Dea     string  // DEA
    Ncp     string  // ncpdp
    Npi     string  // NPI
    Guid    string  // submission_guid
    Winv    string  // wholesaler_invoice_number
    Wacp    string  // wac_price
    Qty     string  // quantity
    Spiq    string  // service_provider_id_qualifier
    Priq    string  // prescriber_id_qualifier
    Prid    string  // prescriber_id
    Netw    string  // network
    Chnm    string  // chain_name
    Prnm    string  // product_name
    Cnfm    string  // claim_conforms_flag
    Phtp    string  // Pharmacy type (1CP, WO, NCP, ...)
    Stor    string  // store ID
    Lbn     string  // legal_business_name
    Name    string  // name
    Isih    string  // Entity has in-house pharmacy - derived by checking in_house_pharmacy_ids to see if it has any values
    Isgr    string  // Entity is grantee
    Dupl    string  // duplicate
    Stge    string  // staging
    Excl    string  // exclusion reason
    Shid    string  // Short unique identifier string used as a claim-id for customers to see.
    Elig    string  // eligible_at_submission
    Susp    string  // suspended_at_submission
    UpAt    string  // updated_at
    Hash    string  // a hash value - typically used to represent the commit hash of the client binary
    Desc    string  // a description - typically used to represent the description of the client binary
    Data    string  // the name of a data source - binary sends to server to indicate table or source to be read/returned.
    Strt    string  // start date/time
    Term    string  // Termination date/time
    State   string  // state
    Inhs    string  // in_house_pharmacy_ids - the array of ids within this column
    Locn    string  // Location (general use)
    ErrC    string
    ErrM    string
    Ecsl    string // ECS lock to set
    Id      string
    Spmt    string  // service provider match type (direct, xwalk, stack, chain)

    Scid   string   // scrub_id
    Clid   string   // claim_id

    Hpid   string   // hcpcs_id
    Pdsc   string   // product_description
    Pos    string   // place_of_service_code
    Hpm1   string   // hcpcs_modifier_1
    Hpm2   string   // hcpcs_modifier_2
    Hpm3   string   // hcpcs_modifier_3
    Hpm4   string   // hcpcs_modifier_4
    Mrst   string   // MRS_Status
}{
    "stat",
    "auth",
    "vers",
    "manu",
    "envr",
    "spid",
    "rxn",
    "frxn",
    "dos",
    "dof",
    "doc",
    "dol",
    "dpr",
    "ndc",
    "i340",
    "istc",
    "ceid",
    "phid",
    "dea",
    "ncp",
    "npi",
    "guid",
    "winv",
    "wacp",
    "qty",
    "spiq",
    "priq",
    "prid",
    "netw",
    "chnm",
    "prnm",
    "cnfm",
    "phtp",
    "stor",
    "lbn",
    "name",
    "isih",
    "isgr",
    "dupl",
    "stge",
    "excl",
    "shid",
    "elig",
    "susp",
    "upat",
    "hash",
    "desc",
    "data",
    "strt",
    "term",
    "state",
    "inhs",
    "locn",
    "errc",
    "errm",
    "ecsl",
    "id",
    "spmt",

    "scid",
    "clid",

    "hpid",
    "pdsc",
    "pos",
    "hpm1",
    "hpm2",
    "hpm3",
    "hpm4",
    "mrst",
}

var FullToShort = map[string]string{
    "status":                           Fields.Stat,
    "manufacturer":                     Fields.Manu,
    "manufacturer_name":                Fields.Manu,
    "manu":                             Fields.Manu,
    "environment":                      Fields.Envr,
    "service_provider_id":              Fields.Spid,
    "rx_number":                        Fields.Rxn,
    "formatted_rx_number":              Fields.Frxn,
    "date_of_service":                  Fields.Dos,
    "date_of_fill":                      Fields.Dof,
    "date_prescribed":                  Fields.Dpr,
    "rbt_hdos":                         Fields.Dol,
    "item":                             Fields.Ndc,
    "id_340b":                          Fields.I340,
    "status_code_340b":                 Fields.Istc,
    "contracted_entity_id":             Fields.Ceid,
    "pharmacy_id":                      Fields.Phid,
    "dea":                              Fields.Dea,
    "dea_registration_id":              Fields.Dea,
    "ncpdp":                            Fields.Ncp,
    "ncpdp_provider_id":                Fields.Ncp,
    "npi":                              Fields.Npi,
    "national_provider_id":             Fields.Npi,
    "submission_guid":                  Fields.Guid,
    "wholesaler_invoice_number":        Fields.Winv,
    "wac_price":                        Fields.Wacp,
    "quantity":                         Fields.Qty,
    "service_provider_id_qualifier":     Fields.Spiq,
    "prescriber_id_qualifier":           Fields.Priq,
    "prescriber_id":                    Fields.Prid,
    "network":                          Fields.Netw,
    "chain_name":                       Fields.Chnm,
    "product_name":                     Fields.Prnm,
    "claim_conforms_flag":               Fields.Cnfm,
    "pharmacy_type":                    Fields.Phtp,
    "store_number":                     Fields.Stor,
    "legal_business_name":              Fields.Lbn,
    "short_id":                         Fields.Shid,
    "eligible_at_submission":           Fields.Elig,
    "suspended_submission":             Fields.Susp,
    "created_at":                       Fields.Doc,
    "updated_at":                       Fields.UpAt,
    "claim_id":                         Fields.Clid,
    "id":                               Fields.Id,
    "sp_match":                         Fields.Spmt,
    "start":                            Fields.Strt,
    "term":                             Fields.Term,
    "state":                            Fields.State,
    "reason":                           Fields.Excl,
    "in_house_pharmacy_ids":            Fields.Inhs,
    "pharmacy_state":                   Fields.State,
    "duplicate":                        Fields.Dupl,

    "hcpcs_id":                         Fields.Hpid,
}

var ShortToFull = map[string]string{
    Fields.Stat:    "status",
    Fields.Manu:    "manufacturer",
    Fields.Envr:    "environment",
    Fields.Spid:    "service_provider_id",
    Fields.Rxn:     "rx_number",
    Fields.Frxn:    "formatted_rx_number",
    Fields.Dos:     "date_of_service",
    Fields.Dof:     "date_of_fill",
    Fields.Dpr:     "date_prescribed",
    Fields.Dol:     "rbt_hdos",
    Fields.Doc:     "created_at",
    Fields.Ndc:     "item",
    Fields.I340:    "id_340b",
    Fields.Istc:    "status_code_340b",
    Fields.Ceid:    "contracted_entity_id",
    Fields.Phid:    "pharmacy_id",
    Fields.Dea:     "dea_registration_id",
    Fields.Ncp:     "ncpdp_provider_id",
    Fields.Npi:     "national_provider_id",
    Fields.Guid:    "submission_guid",
    Fields.Winv:    "wholesaler_invoice_number",
    Fields.Wacp:    "wac_price",
    Fields.Qty:     "quantity",
    Fields.Spiq:    "service_provider_id_qualifier",
    Fields.Priq:    "prescriber_id_qualifier",
    Fields.Prid:    "prescriber_id",
    Fields.Netw:    "network",
    Fields.Chnm:    "chain_name",
    Fields.Prnm:    "product_name",
    Fields.Cnfm:    "claim_conforms_flag",
    Fields.Phtp:    "pharmacy_type",
    Fields.Stor:    "store_number",
    Fields.Lbn:     "legal_business_name",
    Fields.Shid:    "short_id",
    Fields.Elig:    "eligible_at_submission",
    Fields.Susp:    "suspended_submission",
    Fields.UpAt:    "updated_at",
    Fields.Clid:    "claim_id",
    Fields.Strt:    "start",
    Fields.Term:    "term",
    Fields.State:   "state",
    Fields.Excl:    "reason",
    Fields.Dol:     "rbt_hdos",
    Fields.Inhs:    "in_house_pharmacy_ids",
    Fields.Dupl:    "duplicate",
    Fields.Id:      "id",
    Fields.Spmt:    "sp_match",

    Fields.Hpid:    "hcpcs_id",
}

func ToShortName(full string) string {
    if short,ok := FullToShort[full];ok {
        return short
    } else if _,ok := ShortToFull[full];ok {
        // Already a short name. Just return it back.
        return full
    }
    return "" // Not a full name nor short name.
}
func ToFullName(short string) string {
    if full,ok := ShortToFull[short];ok {
        return full
    } else if _,ok := FullToShort[short];ok {
        // Already a full name. Just return it back.
        return short
    }
    return "" // Not a full name nor short name.
}
