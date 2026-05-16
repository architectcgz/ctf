CREATE TABLE public.awd_scope_controls (
    id bigint PRIMARY KEY,
    contest_id bigint NOT NULL,
    team_id bigint NOT NULL,
    scope_type character varying(24) NOT NULL,
    service_id bigint NOT NULL DEFAULT 0,
    control_type character varying(48) NOT NULL,
    reason text NOT NULL DEFAULT '',
    updated_by bigint,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT awd_scope_controls_contest_id_fkey FOREIGN KEY (contest_id) REFERENCES public.contests(id) ON DELETE CASCADE,
    CONSTRAINT awd_scope_controls_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(id) ON DELETE CASCADE,
    CONSTRAINT awd_scope_controls_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES public.users(id) ON DELETE SET NULL,
    CONSTRAINT chk_awd_scope_controls_scope_type CHECK (scope_type IN ('team', 'team_service')),
    CONSTRAINT chk_awd_scope_controls_control_type CHECK (control_type IN ('retired', 'service_disabled', 'desired_reconcile_suppressed')),
    CONSTRAINT chk_awd_scope_controls_team_scope_service_id CHECK (
        (scope_type = 'team' AND service_id = 0) OR
        (scope_type = 'team_service' AND service_id > 0)
    )
);

CREATE SEQUENCE public.awd_scope_controls_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.awd_scope_controls_id_seq OWNED BY public.awd_scope_controls.id;

ALTER TABLE ONLY public.awd_scope_controls
    ALTER COLUMN id SET DEFAULT nextval('public.awd_scope_controls_id_seq'::regclass);

CREATE INDEX idx_awd_scope_controls_scope
ON public.awd_scope_controls USING btree (contest_id, team_id, scope_type, service_id);

CREATE UNIQUE INDEX uk_awd_scope_controls
ON public.awd_scope_controls USING btree (contest_id, team_id, scope_type, service_id, control_type);
