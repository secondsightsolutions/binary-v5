
CREATE TABLE atlas.scrubs (
    scid bigserial primary key,
    auth text not null,
    plcy text not null,
    name text not null,
    vers text not null,
    dscr text not null,
    hash text not null,
    host text not null,
    appl text not null,
    hdrs text not null,
    cmdl text not null,
    created timestamp not null default now(),
    ready   timestamp,
    started timestamp,
    done    timestamp
);

CREATE TABLE atlas.rebates (
    seq   bigserial,
    scid  bigint not null,
    rbid  bigint not null,
    indx  bigint not null default 0,
    rxn   text not null default '',
    hrxn  text not null default '',
    ndc   text not null default '',
    spid  text not null default '',
    prid  text not null default '',
    dos   text not null default '',
    stat  text not null default '',
    excl  text not null default '',
    errc  text not null default '',
    errm  text not null default '',
    col1  text not null default '',
    col2  text not null default '',
    col3  text not null default '',
    col4  text not null default '',
    col5  text not null default '',
    col6  text not null default '',
    col7  text not null default '',
    col8  text not null default '',
    col9  text not null default '',
    col10 text not null default '',
    CONSTRAINT rebates_pk PRIMARY KEY (scid, rbid)
);
CREATE INDEX ON atlas.rebates(seq);
CREATE INDEX ON atlas.rebates(scid);
CREATE INDEX ON atlas.rebates(scid, indx);

CREATE TABLE atlas.claim_uses (
    seq   bigserial,
    scid  bigint not null,
    shrt  text not null,
    excl  text not null default '',
    CONSTRAINT claim_uses_pk PRIMARY KEY (scid, shrt)
);
CREATE INDEX ON atlas.claim_uses(seq);
CREATE INDEX ON atlas.claim_uses(scid);
CREATE INDEX ON atlas.claim_uses(shrt);
CREATE INDEX ON atlas.claim_uses(scid, shrt);

CREATE TABLE atlas.rebate_meta (
    seq   bigserial,
    scid  bigint not null primary key,
    col1  text not null default '',
    col2  text not null default '',
    col3  text not null default '',
    col4  text not null default '',
    col5  text not null default '',
    col6  text not null default '',
    col7  text not null default '',
    col8  text not null default '',
    col9  text not null default '',
    col10 text not null default ''
);
CREATE INDEX ON atlas.rebate_meta(seq);
CREATE INDEX ON atlas.rebate_meta(scid);

CREATE TABLE atlas.rebate_claims (
    seq  bigserial,
    scid bigint not null,
    rbid bigint not null,
    shrt text not null,
    CONSTRAINT rebate_claims_pk PRIMARY KEY (scid, rbid, shrt)
);
CREATE INDEX ON atlas.rebate_claims(seq);

CREATE TABLE atlas.sync (
    claims        timestamp not null default '2000-01-01',
    scrubs        bigint not null default 0,
    rebates       bigint not null default 0,
    claim_uses    bigint not null default 0,
    rebate_meta   bigint not null default 0,
    rebate_claims bigint not null default 0
);

CREATE TABLE atlas.claims (
    shrt text primary key,
    i340 text not null,
    ndc  text not null,
    spid text not null,
    prid text not null default '',
    hrxn text not null,
    hfrx text not null default '',
    hdos text not null,
    hdop text not null,
    doc  bigint not null default 0,
    dos  bigint not null default 0,
    dop  bigint not null default 0,
    netw text not null,
    prnm text not null,
    chnm text not null default '',
    elig bool not null default true,
    susp bool not null default false,
    ihph text array not null default '{}'
);
