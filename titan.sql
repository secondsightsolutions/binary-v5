
DROP TABLE IF EXISTS titan.scrub_rebates_claims;
DROP TABLE IF EXISTS titan.scrub_matches;
DROP TABLE IF EXISTS titan.scrub_rebates;
DROP TABLE IF EXISTS titan.scrub_claims;
DROP TABLE IF EXISTS titan.metrics;
DROP TABLE IF EXISTS titan.scrubs;
DROP TABLE IF EXISTS titan.claims;
DROP TABLE IF EXISTS titan.requests;
DROP TABLE IF EXISTS titan.auth;
DROP TABLE IF EXISTS titan.commands;
DROP TABLE IF EXISTS titan.desigs;
DROP TABLE IF EXISTS titan.eligibility;
DROP TABLE IF EXISTS titan.entities;
DROP TABLE IF EXISTS titan.esp1;
DROP TABLE IF EXISTS titan.ldns;
DROP TABLE IF EXISTS titan.ndcs;
DROP TABLE IF EXISTS titan.pharmacies;
DROP TABLE IF EXISTS titan.spis;

CREATE TABLE titan.requests (
    rqid bigserial primary key,
    comd text not null,
    manu text not null,
    name text not null,
    auth text not null,
    vers text not null,
    xou  text not null,
    dscr text not null,
    hash text not null,
    netw text not null default '',
    addr text not null default '',
    host text not null default '',
    rslt text not null default '',
    cmid bigint not null default -1,
    crat timestamp with time zone not null default now()
);

CREATE TABLE titan.commands (
    cmid bigint not null,
    comd text not null,
    manu text not null,
    name text not null,
    auth text not null,
    vers text not null,
    kind text not null default '',
    xou  text not null,
    dscr text not null,
    hash text not null,
    netw text not null default '',
    host text not null default '',
    addr text not null default '',
    cwd  text not null default '',
    "user" text not null default '',
    cmdl text not null,
    rslt text not null default '',
    crat timestamp with time zone not null default now(), -- created
    seq  bigint not null,
    CONSTRAINT commands_pk PRIMARY KEY (manu, cmid)
);
CREATE INDEX ON atlas.commands(seq);

CREATE TABLE titan.scrubs (
    manu text not null,
    scid bigint not null,
    ivid bigint not null,
    cmid bigint not null,
    plcy text not null,
    hdrs text not null,
    crat timestamp with time zone not null, -- created
    rdat timestamp with time zone, -- ready
    srat timestamp with time zone, -- started
    dnat timestamp with time zone, -- done
    seq  bigint not null,
    CONSTRAINT scrubs_pk PRIMARY KEY (manu, scid)
);

CREATE TABLE titan.metrics (
    scid 				bigint not null,
    manu                text not null,
    rbt_total           integer not null default 0,
    rbt_matched         integer not null default 0,
    rbt_nomatch         integer not null default 0,
    rbt_invalid         integer not null default 0,
    rbt_passed          integer not null default 0,
    rbt_failed          integer not null default 0,
    clm_total           integer not null default 0,
    clm_valid           integer not null default 0,
    clm_matched         integer not null default 0,
    clm_nomatch         integer not null default 0,
    clm_invalid         integer not null default 0,
    spi_exact           integer not null default 0,
    spi_cross           integer not null default 0,
    spi_stack           integer not null default 0,
    spi_chain           integer not null default 0,
    dos_equ_doc         integer not null default 0,
    dos_bef_doc         integer not null default 0,
    dos_aft_doc         integer not null default 0,
    dos_equ_dof         integer not null default 0,
    dos_bef_dof         integer not null default 0,
    dos_aft_dof         integer not null default 0,
    dos_range_pass      integer not null default 0,
    dos_range_fail      integer not null default 0,
    dof_range_pass      integer not null default 0,
    dof_range_fail      integer not null default 0,
    doc_range_pass      integer not null default 0,
    doc_range_fail      integer not null default 0,
    r_no_match_rx       integer not null default 0,
    r_no_match_spi      integer not null default 0,
    r_no_match_ndc      integer not null default 0,
    r_clm_used          integer not null default 0,
    r_dos_rbt_aft_clm   integer not null default 0,
    r_dos_clm_aft_rbt   integer not null default 0,
    r_old_rbt_new_clm   integer not null default 0,
    r_new_rbt_old_clm   integer not null default 0,
    r_clm_not_cnfm      integer not null default 0,
    r_phm_not_desg      integer not null default 0,
    r_inv_desg_type     integer not null default 0,
    r_wrong_network     integer not null default 0,
    r_not_elig_at_sub   integer not null default 0,
    seq                 bigint not null,
    CONSTRAINT metrics_pk PRIMARY KEY (manu, scid),
    FOREIGN KEY (manu, scid) references titan.scrubs(manu, scid)
);

CREATE TABLE titan.claims (
	manu text not null,
    clid text not null,
    i340 text not null,
    ndc  text not null,
    spid text not null,
    prid text not null default '',
    hrxn text not null,
    hfrx text not null default '',
    hdos text not null,
    hdop text not null,
    doc  timestamp with time zone,
    dos  timestamp with time zone,
    dop  timestamp with time zone,
    netw text not null,
    prnm text not null,
    chnm text not null default '',
    elig bool not null default true,
    susp bool not null default false,
    cnfm bool not null default true,
    qty  numeric not null default 0,
    ihph text not null default '',
    seq  bigint not null,
    CONSTRAINT claims_pk PRIMARY KEY(manu, clid)
);
CREATE INDEX ON titan.claims(manu);
CREATE INDEX ON titan.claims(manu, doc);
CREATE INDEX ON titan.claims(manu, seq);

CREATE TABLE titan.scrub_rebates (
    manu text not null,
    scid bigint not null,
    rbid bigint not null,
    indx bigint not null default 0,
    stat text not null default '',
    excl text not null default '',
    spmt text not null default '',
    fprt text not null default '',
    seq  bigint not null,
    CONSTRAINT scrub_rebates_pk PRIMARY KEY (manu, scid, rbid) --,
    -- FOREIGN KEY (manu, scid) references titan.scrubs(manu, scid)
);
CREATE INDEX ON titan.scrub_rebates(manu);
CREATE INDEX ON titan.scrub_rebates(manu, scid);

CREATE TABLE titan.scrub_claims (
    manu  text not null,
    scid  bigint not null,
    clid  text not null,
    excl  text not null default '',
    seq   bigint not null,
    CONSTRAINT scrub_claims_pk PRIMARY KEY (manu, scid, clid)
);
CREATE INDEX ON titan.scrub_claims(scid);
CREATE INDEX ON titan.scrub_claims(clid);

CREATE TABLE titan.scrub_matches (
    manu text not null,
    scid bigint not null,
    ivid bigint not null,
    rbid bigint not null,
    clid text not null,
    seq  bigint not null,
    CONSTRAINT scrub_matches_pk PRIMARY KEY (manu, scid, rbid, clid)
);
CREATE TABLE titan.scrub_attempts (
    manu  text not null,
    scid bigint not null,
    ivid bigint not null,
    rbid bigint not null,
    clid text   not null,
    excl text   not null,
    seq  bigserial,
    CONSTRAINT scrub_attempts_pk PRIMARY KEY (manu, scid, rbid, clid)
);

CREATE TABLE titan.auth (
    manu text not null,
	proc text not null,
    auth text not null,
    kind text not null default 'pharmacy',
	vers int8 not null default 0,
    enb  bool not null default true,
	CONSTRAINT auth_pkey PRIMARY KEY (manu, proc, auth, kind)
);

CREATE TABLE titan.entities (
    i340 text   not null,
    strt date,
    term date,
    stat text   not null default '',
    seq  bigint not null primary key
);
CREATE INDEX ON titan.entities(seq);

CREATE TABLE titan.pharmacies (
    i340 text not null,
    phid text not null,
    ncps text not null default '',
    npis text not null default '',
    deas text not null default '',
    chnm text not null default '',
    stat text not null default '',
    seq  bigint not null primary key
);

CREATE TABLE titan.ndcs (
    manu text not null,
    ndc  text not null,
    name text not null,
    netw text not null default 'retail',
    seq  bigint not null,
    CONSTRAINT ndcs_pk PRIMARY KEY (manu, ndc)
);
CREATE INDEX ON titan.ndcs(manu);
CREATE INDEX ON titan.ndcs(seq);

CREATE TABLE titan.spis (
    ncp  text not null primary key,
    npi  text not null default '',
    dea  text not null default '',
    sto  text not null default '',
    nam  text not null default '',
    lbn  text not null default '',
    chn  text not null default '',
    cde  text not null default '',
    seq  bigint not null
);
CREATE INDEX ON titan.spis(seq);

CREATE TABLE titan.desigs (
    manu text not null,
    i340 text not null,
    phid text not null,
    netw text not null default 'retail',
    flag text not null default '',
    hin  text not null default '',
    assg boolean not null default true,
    term boolean not null default false,
    excl boolean not null default false,
    xdat timestamp with time zone,
    dlat timestamp with time zone,
    xsat timestamp with time zone,
    crat timestamp with time zone,
    cpat timestamp with time zone,
    seq  bigint not null,
    CONSTRAINT desigs_pk PRIMARY KEY (manu, i340, phid)
);
CREATE INDEX ON titan.desigs(manu);

CREATE TABLE titan.ldns (
    manu text not null,
    netw text not null,
    phid text not null,
    assg boolean not null default true,
    term boolean not null default false,
    seq  bigint not null,
    CONSTRAINT ldns_pk PRIMARY KEY (manu, netw, phid)
);
CREATE INDEX ON titan.ldns(manu);

CREATE TABLE titan.esp1 (
    manu text   not null,
    spid text   not null,
    ndc  text   not null,
    strt date,
    term date,
    CONSTRAINT esp1_pk PRIMARY KEY (manu, spid, ndc)
);

CREATE TABLE titan.eligibility (
    seq  bigint not null primary key,
    manu text   not null,
    i340 text   not null,
    phid text   not null,
    netw text   not null default 'retail',
    strt timestamp with time zone,
    term timestamp with time zone
);
CREATE INDEX ON titan.eligibility(manu);
