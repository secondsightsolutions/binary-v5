
CREATE TABLE titan.scrubs (
    manu text not null,
    scid bigint not null,
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
    rbt_total int not null default 0,
    rbt_valid int not null default 0,
    rbt_matched int not null default 0,
    rbt_nomatch int not null default 0,
    rbt_passed  int not null default 0,
    rbt_failed  int not null default 0,
    clm_total   int not null default 0,
    clm_valid   int not null default 0,
    clm_matched int not null default 0,
    clm_nomatch int not null default 0,
    clm_invalid int not null default 0,
    spi_exact   int not null default 0,
    spi_cross   int not null default 0,
    spi_stack   int not null default 0,
    spi_chain   int not null default 0,
    dos_equ_doc int not null default 0,
    dos_bef_doc int not null default 0,
    dos_equ_dof int not null default 0,
    dos_bef_dof int not null default 0,
    dos_aft_dof int not null default 0,
    CONSTRAINT PRIMARY KEY (manu, scid)
);

CREATE TABLE titan.rebates (
    manu text not null,
    scid bigint not null,
    rbid bigint not null,
    stat text not null default '',
    fprt text not null default '',
    CONSTRAINT PRIMARY KEY (manu, scid, rbid)
);
CREATE INDEX ON titan.rebates(manu);
CREATE INDEX ON titan.rebates(manu, scid);

CREATE TABLE titan.claim_uses (
    scid  bigint not null,
    shrt  text not null,
    excl  text not null default '',
    CONSTRAINT PRIMARY KEY (scid, shrt)
);
CREATE INDEX ON titan.claim_uses(seq);
CREATE INDEX ON titan.claim_uses(scid);
CREATE INDEX ON titan.claim_uses(shrt);
CREATE INDEX ON titan.claim_uses(scid, shrt);

CREATE TABLE titan.rebate_meta (
    scid  bigint not null primary key,
    col1  text not null default '',
    col2  text not null default '',
    col50 text not null default ''
);
CREATE INDEX ON titan.rebate_meta(seq);
CREATE INDEX ON titan.rebate_meta(scid);

CREATE TABLE titan.rebate_claims (
    scid bigint not null,
    rbid bigint not null,
    shrt text not null,
    CONSTRAINT PRIMARY KEY (manu, scid, rbid, shrt)
);