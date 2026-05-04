--
-- PostgreSQL database dump
--


-- Dumped from database version 16.13
-- Dumped by pg_dump version 16.13

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--



--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: -
--



SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.audit_logs (
    id bigint NOT NULL,
    user_id bigint,
    action character varying(32) NOT NULL,
    resource_type character varying(64) NOT NULL,
    resource_id bigint,
    detail jsonb DEFAULT '{}'::jsonb NOT NULL,
    ip_address character varying(45) NOT NULL,
    user_agent character varying(512) DEFAULT NULL::character varying,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: audit_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.audit_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: audit_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.audit_logs_id_seq OWNED BY public.audit_logs.id;


--
-- Name: awd_attack_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.awd_attack_logs (
    id bigint NOT NULL,
    round_id bigint NOT NULL,
    attacker_team_id bigint NOT NULL,
    victim_team_id bigint NOT NULL,
    awd_challenge_id bigint NOT NULL,
    attack_type character varying(32) NOT NULL,
    submitted_flag character varying(512) DEFAULT NULL::character varying,
    is_success boolean DEFAULT false NOT NULL,
    score_gained integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    source character varying(32) DEFAULT 'legacy'::character varying NOT NULL,
    submitted_by_user_id bigint,
    service_id bigint NOT NULL
);


--
-- Name: awd_attack_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.awd_attack_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: awd_attack_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.awd_attack_logs_id_seq OWNED BY public.awd_attack_logs.id;


--
-- Name: awd_challenges; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.awd_challenges (
    id bigint NOT NULL,
    name character varying(128) NOT NULL,
    slug character varying(128) NOT NULL,
    category character varying(64) DEFAULT ''::character varying NOT NULL,
    difficulty character varying(32) DEFAULT ''::character varying NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    service_type character varying(32) NOT NULL,
    deployment_mode character varying(32) NOT NULL,
    version character varying(32) DEFAULT 'v1'::character varying NOT NULL,
    status character varying(24) DEFAULT 'draft'::character varying NOT NULL,
    checker_type character varying(32) DEFAULT ''::character varying NOT NULL,
    checker_config text DEFAULT '{}'::text NOT NULL,
    flag_mode character varying(32) DEFAULT ''::character varying NOT NULL,
    flag_config text DEFAULT '{}'::text NOT NULL,
    defense_entry_mode character varying(32) DEFAULT ''::text NOT NULL,
    access_config text DEFAULT '{}'::text NOT NULL,
    runtime_config text DEFAULT '{}'::text NOT NULL,
    readiness_status character varying(24) DEFAULT 'pending'::character varying NOT NULL,
    readiness_report text DEFAULT ''::text NOT NULL,
    last_verified_at timestamp without time zone,
    last_verified_by bigint,
    created_by bigint,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp without time zone
);


--
-- Name: awd_challenges_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.awd_challenges_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: awd_challenges_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.awd_challenges_id_seq OWNED BY public.awd_challenges.id;


--
-- Name: awd_rounds; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.awd_rounds (
    id bigint NOT NULL,
    contest_id bigint NOT NULL,
    round_number integer NOT NULL,
    status character varying(16) DEFAULT 'pending'::character varying NOT NULL,
    started_at timestamp with time zone,
    ended_at timestamp with time zone,
    attack_score integer DEFAULT 50 NOT NULL,
    defense_score integer DEFAULT 50 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: awd_rounds_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.awd_rounds_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: awd_rounds_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.awd_rounds_id_seq OWNED BY public.awd_rounds.id;


--
-- Name: awd_service_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.awd_service_templates (
    id bigint NOT NULL,
    name character varying(128) NOT NULL,
    slug character varying(128) NOT NULL,
    category character varying(64) DEFAULT ''::character varying NOT NULL,
    difficulty character varying(32) DEFAULT ''::character varying NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    service_type character varying(32) NOT NULL,
    deployment_mode character varying(32) NOT NULL,
    version character varying(32) DEFAULT 'v1'::character varying NOT NULL,
    status character varying(24) DEFAULT 'draft'::character varying NOT NULL,
    checker_type character varying(32) DEFAULT ''::character varying NOT NULL,
    checker_config text DEFAULT '{}'::text NOT NULL,
    flag_mode character varying(32) DEFAULT ''::character varying NOT NULL,
    flag_config text DEFAULT '{}'::text NOT NULL,
    defense_entry_mode character varying(32) DEFAULT ''::character varying NOT NULL,
    access_config text DEFAULT '{}'::text NOT NULL,
    runtime_config text DEFAULT '{}'::text NOT NULL,
    readiness_status character varying(24) DEFAULT 'pending'::character varying NOT NULL,
    readiness_report text DEFAULT ''::text NOT NULL,
    last_verified_at timestamp without time zone,
    last_verified_by bigint,
    created_by bigint,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp without time zone
);


--
-- Name: awd_service_templates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.awd_service_templates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: awd_service_templates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.awd_service_templates_id_seq OWNED BY public.awd_service_templates.id;


--
-- Name: awd_team_services; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.awd_team_services (
    id bigint NOT NULL,
    round_id bigint NOT NULL,
    team_id bigint NOT NULL,
    awd_challenge_id bigint NOT NULL,
    service_status character varying(16) DEFAULT 'up'::character varying NOT NULL,
    check_result jsonb DEFAULT '{}'::jsonb NOT NULL,
    attack_received integer DEFAULT 0 NOT NULL,
    defense_score integer DEFAULT 0 NOT NULL,
    attack_score integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    sla_score integer DEFAULT 0 NOT NULL,
    checker_type character varying(32) DEFAULT ''::character varying NOT NULL,
    service_id bigint NOT NULL
);


--
-- Name: awd_team_services_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.awd_team_services_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: awd_team_services_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.awd_team_services_id_seq OWNED BY public.awd_team_services.id;


--
-- Name: awd_traffic_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.awd_traffic_events (
    id bigint NOT NULL,
    contest_id bigint NOT NULL,
    round_id bigint NOT NULL,
    attacker_team_id bigint NOT NULL,
    victim_team_id bigint NOT NULL,
    awd_challenge_id bigint NOT NULL,
    method character varying(16) NOT NULL,
    path character varying(1024) NOT NULL,
    status_code integer NOT NULL,
    source character varying(32) DEFAULT 'runtime_proxy'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    service_id bigint NOT NULL
);


--
-- Name: awd_traffic_events_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.awd_traffic_events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: awd_traffic_events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.awd_traffic_events_id_seq OWNED BY public.awd_traffic_events.id;


--
-- Name: challenge_hints; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.challenge_hints (
    id bigint NOT NULL,
    challenge_id bigint,
    level bigint,
    title text,
    content text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: challenge_hints_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.challenge_hints_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: challenge_hints_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.challenge_hints_id_seq OWNED BY public.challenge_hints.id;


--
-- Name: challenge_package_revisions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.challenge_package_revisions (
    id bigint NOT NULL,
    challenge_id bigint NOT NULL,
    revision_no integer NOT NULL,
    source_type character varying(32) NOT NULL,
    parent_revision_id bigint,
    package_slug character varying(128) NOT NULL,
    archive_path text DEFAULT ''::text NOT NULL,
    source_dir text DEFAULT ''::text NOT NULL,
    manifest_snapshot text DEFAULT ''::text NOT NULL,
    topology_source_path character varying(255) DEFAULT ''::character varying NOT NULL,
    topology_snapshot text DEFAULT ''::text NOT NULL,
    created_by bigint,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: challenge_package_revisions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.challenge_package_revisions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: challenge_package_revisions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.challenge_package_revisions_id_seq OWNED BY public.challenge_package_revisions.id;


--
-- Name: challenge_publish_check_jobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.challenge_publish_check_jobs (
    id bigint NOT NULL,
    challenge_id bigint NOT NULL,
    requested_by bigint NOT NULL,
    status character varying(32) NOT NULL,
    request_source character varying(32) DEFAULT 'admin_publish'::character varying NOT NULL,
    result_json text DEFAULT ''::text NOT NULL,
    failure_summary character varying(512) DEFAULT ''::character varying NOT NULL,
    published_at timestamp with time zone,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: challenge_publish_check_jobs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.challenge_publish_check_jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: challenge_publish_check_jobs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.challenge_publish_check_jobs_id_seq OWNED BY public.challenge_publish_check_jobs.id;


--
-- Name: challenge_tags; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.challenge_tags (
    id bigint NOT NULL,
    challenge_id bigint NOT NULL,
    tag_id bigint NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: challenge_tags_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.challenge_tags_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: challenge_tags_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.challenge_tags_id_seq OWNED BY public.challenge_tags.id;


--
-- Name: challenge_topologies; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.challenge_topologies (
    id bigint NOT NULL,
    challenge_id bigint NOT NULL,
    template_id bigint,
    entry_node_key character varying(64) NOT NULL,
    spec jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    source_type character varying(32) DEFAULT 'platform_manual'::character varying NOT NULL,
    source_path character varying(255) DEFAULT ''::character varying NOT NULL,
    package_revision_id bigint,
    package_baseline_spec jsonb DEFAULT '{}'::jsonb NOT NULL,
    sync_status character varying(32) DEFAULT 'clean'::character varying NOT NULL,
    last_export_revision_id bigint
);


--
-- Name: challenge_topologies_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.challenge_topologies_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: challenge_topologies_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.challenge_topologies_id_seq OWNED BY public.challenge_topologies.id;


--
-- Name: challenge_writeups; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.challenge_writeups (
    id bigint NOT NULL,
    challenge_id bigint NOT NULL,
    title character varying(256) NOT NULL,
    content text NOT NULL,
    visibility character varying(16) DEFAULT 'private'::character varying NOT NULL,
    created_by bigint,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    is_recommended boolean DEFAULT false NOT NULL,
    recommended_at timestamp with time zone,
    recommended_by bigint
);


--
-- Name: challenge_writeups_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.challenge_writeups_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: challenge_writeups_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.challenge_writeups_id_seq OWNED BY public.challenge_writeups.id;


--
-- Name: challenges; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.challenges (
    id bigint NOT NULL,
    title text,
    description text,
    category text,
    difficulty text,
    points bigint,
    image_id bigint,
    status text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    flag_prefix character varying(32) DEFAULT 'flag'::character varying,
    flag_type character varying(16) DEFAULT 'static'::character varying NOT NULL,
    flag_hash character varying(128),
    flag_salt character varying(64),
    attachment_url text,
    created_by bigint,
    package_slug character varying(128),
    flag_regex character varying(512),
    instance_sharing character varying(16) DEFAULT 'per_user'::character varying NOT NULL,
    target_protocol character varying(16) DEFAULT 'http'::character varying NOT NULL,
    target_port integer DEFAULT 0 NOT NULL,
    CONSTRAINT chk_challenges_instance_sharing CHECK (((instance_sharing)::text = ANY (ARRAY[('per_user'::character varying)::text, ('per_team'::character varying)::text, ('shared'::character varying)::text]))),
    CONSTRAINT chk_challenges_target_protocol CHECK (((target_protocol)::text = ANY ((ARRAY['http'::character varying, 'tcp'::character varying])::text[])))
);


--
-- Name: challenges_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.challenges_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: challenges_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.challenges_id_seq OWNED BY public.challenges.id;


--
-- Name: contest_announcements; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.contest_announcements (
    id bigint NOT NULL,
    contest_id bigint NOT NULL,
    title character varying(200) NOT NULL,
    content text DEFAULT ''::text NOT NULL,
    created_by bigint,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: contest_announcements_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.contest_announcements_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: contest_announcements_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.contest_announcements_id_seq OWNED BY public.contest_announcements.id;


--
-- Name: contest_awd_services; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.contest_awd_services (
    id bigint NOT NULL,
    contest_id bigint NOT NULL,
    awd_challenge_id bigint NOT NULL,
    display_name character varying(128) DEFAULT ''::character varying NOT NULL,
    "order" integer DEFAULT 0 NOT NULL,
    is_visible boolean DEFAULT true NOT NULL,
    score_config text DEFAULT '{}'::text NOT NULL,
    runtime_config text DEFAULT '{}'::text NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp without time zone,
    awd_checker_validation_state character varying(24) DEFAULT 'pending'::character varying NOT NULL,
    awd_checker_last_preview_at timestamp without time zone,
    awd_checker_last_preview_result text DEFAULT ''::text NOT NULL,
    service_snapshot text DEFAULT '{}'::text NOT NULL
);


--
-- Name: contest_awd_services_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.contest_awd_services_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: contest_awd_services_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.contest_awd_services_id_seq OWNED BY public.contest_awd_services.id;


--
-- Name: contest_challenges; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.contest_challenges (
    id bigint NOT NULL,
    contest_id bigint NOT NULL,
    challenge_id bigint NOT NULL,
    points integer DEFAULT 0 NOT NULL,
    "order" integer DEFAULT 0 NOT NULL,
    is_visible boolean DEFAULT true NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp without time zone,
    contest_score integer,
    first_blood_by bigint
);


--
-- Name: contest_challenges_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.contest_challenges_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: contest_challenges_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.contest_challenges_id_seq OWNED BY public.contest_challenges.id;


--
-- Name: contest_registrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.contest_registrations (
    id bigint NOT NULL,
    contest_id bigint NOT NULL,
    user_id bigint NOT NULL,
    team_id bigint,
    status character varying(16) DEFAULT 'pending'::character varying NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    reviewed_by bigint,
    reviewed_at timestamp without time zone
);


--
-- Name: contest_registrations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.contest_registrations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: contest_registrations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.contest_registrations_id_seq OWNED BY public.contest_registrations.id;


--
-- Name: contests; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.contests (
    id bigint NOT NULL,
    title character varying(200) NOT NULL,
    description text,
    mode character varying(20) NOT NULL,
    start_time timestamp with time zone NOT NULL,
    end_time timestamp with time zone NOT NULL,
    status character varying(20) DEFAULT 'draft'::character varying NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp without time zone,
    freeze_time timestamp with time zone,
    CONSTRAINT check_time_range CHECK ((end_time > start_time)),
    CONSTRAINT contests_mode_check CHECK (((mode)::text = ANY (ARRAY[('jeopardy'::character varying)::text, ('awd'::character varying)::text]))),
    CONSTRAINT contests_status_check CHECK (((status)::text = ANY (ARRAY[('draft'::character varying)::text, ('registration'::character varying)::text, ('running'::character varying)::text, ('frozen'::character varying)::text, ('ended'::character varying)::text])))
);


--
-- Name: contests_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.contests_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: contests_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.contests_id_seq OWNED BY public.contests.id;


--
-- Name: environment_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.environment_templates (
    id bigint NOT NULL,
    name character varying(128) NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    entry_node_key character varying(64) NOT NULL,
    spec jsonb DEFAULT '{}'::jsonb NOT NULL,
    usage_count integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: environment_templates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.environment_templates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: environment_templates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.environment_templates_id_seq OWNED BY public.environment_templates.id;


--
-- Name: images; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.images (
    id bigint NOT NULL,
    name text,
    tag text,
    description text,
    size bigint,
    status text,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    deleted_at timestamp without time zone
);


--
-- Name: images_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.images_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: images_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.images_id_seq OWNED BY public.images.id;


--
-- Name: instances; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.instances (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    challenge_id bigint NOT NULL,
    container_id character varying(64) NOT NULL,
    network_id character varying(64) DEFAULT NULL::character varying,
    status character varying(16) NOT NULL,
    access_url character varying(255) DEFAULT NULL::character varying,
    expires_at timestamp with time zone NOT NULL,
    extend_count integer DEFAULT 0 NOT NULL,
    max_extends integer DEFAULT 2 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    nonce character varying(64),
    runtime_details text DEFAULT ''::text NOT NULL,
    contest_id bigint,
    team_id bigint,
    host_port integer,
    share_scope character varying(16) DEFAULT 'per_user'::character varying NOT NULL,
    service_id bigint,
    destroyed_at timestamp with time zone,
    CONSTRAINT chk_instances_share_scope CHECK (((share_scope)::text = ANY (ARRAY[('per_user'::character varying)::text, ('per_team'::character varying)::text, ('shared'::character varying)::text])))
);


--
-- Name: instances_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.instances_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: instances_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.instances_id_seq OWNED BY public.instances.id;


--
-- Name: notification_batches; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notification_batches (
    id bigint NOT NULL,
    created_by bigint NOT NULL,
    type character varying(16) NOT NULL,
    title character varying(256) NOT NULL,
    content text NOT NULL,
    link character varying(512) DEFAULT NULL::character varying,
    audience_mode character varying(16) NOT NULL,
    audience_rules jsonb DEFAULT '{}'::jsonb NOT NULL,
    recipient_count integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    published_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notification_batches_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.notification_batches_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: notification_batches_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.notification_batches_id_seq OWNED BY public.notification_batches.id;


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    type character varying(16) NOT NULL,
    title character varying(256) NOT NULL,
    content text NOT NULL,
    is_read boolean DEFAULT false NOT NULL,
    link character varying(512) DEFAULT NULL::character varying,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    read_at timestamp with time zone,
    batch_id bigint
);


--
-- Name: notifications_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.notifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: notifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.notifications_id_seq OWNED BY public.notifications.id;


--
-- Name: port_allocations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.port_allocations (
    port integer NOT NULL,
    instance_id bigint,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: reports; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reports (
    id bigint NOT NULL,
    type character varying(32) NOT NULL,
    format character varying(16) NOT NULL,
    user_id bigint,
    class_name character varying(128) DEFAULT NULL::character varying,
    status character varying(32) NOT NULL,
    file_path text DEFAULT ''::text NOT NULL,
    expires_at timestamp with time zone,
    error_msg text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    completed_at timestamp with time zone
);


--
-- Name: reports_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.reports_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: reports_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.reports_id_seq OWNED BY public.reports.id;


--
-- Name: roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.roles (
    id bigint NOT NULL,
    code character varying(32) NOT NULL,
    name character varying(64) NOT NULL,
    description character varying(256) DEFAULT NULL::character varying,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: roles_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.roles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: roles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.roles_id_seq OWNED BY public.roles.id;


--
-- Name: skill_profiles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.skill_profiles (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    dimension character varying(20) NOT NULL,
    score numeric(5,4) DEFAULT 0 NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT skill_profiles_dimension_check CHECK (((dimension)::text = ANY (ARRAY[('web'::character varying)::text, ('pwn'::character varying)::text, ('reverse'::character varying)::text, ('crypto'::character varying)::text, ('misc'::character varying)::text, ('forensics'::character varying)::text]))),
    CONSTRAINT skill_profiles_score_check CHECK (((score >= (0)::numeric) AND (score <= (1)::numeric)))
);


--
-- Name: skill_profiles_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.skill_profiles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: skill_profiles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.skill_profiles_id_seq OWNED BY public.skill_profiles.id;


--
-- Name: submission_writeups; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.submission_writeups (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    challenge_id bigint NOT NULL,
    contest_id bigint,
    title character varying(256) NOT NULL,
    content text NOT NULL,
    submission_status character varying(16) DEFAULT 'draft'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    visibility_status character varying(16) DEFAULT 'visible'::character varying NOT NULL,
    is_recommended boolean DEFAULT false NOT NULL,
    recommended_at timestamp with time zone,
    recommended_by bigint,
    published_at timestamp with time zone
);


--
-- Name: submission_writeups_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.submission_writeups_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: submission_writeups_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.submission_writeups_id_seq OWNED BY public.submission_writeups.id;


--
-- Name: submissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.submissions (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    challenge_id bigint NOT NULL,
    flag character varying(500),
    is_correct boolean NOT NULL,
    submitted_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    contest_id bigint,
    team_id bigint,
    score integer DEFAULT 0 NOT NULL,
    review_status character varying(32) DEFAULT 'not_required'::character varying NOT NULL,
    reviewed_by bigint,
    reviewed_at timestamp with time zone,
    review_comment text DEFAULT ''::text NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: submissions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.submissions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: submissions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.submissions_id_seq OWNED BY public.submissions.id;


--
-- Name: tags; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tags (
    id bigint NOT NULL,
    name character varying(64) NOT NULL,
    type character varying(32) DEFAULT 'vulnerability'::character varying NOT NULL,
    description text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: tags_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tags_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tags_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tags_id_seq OWNED BY public.tags.id;


--
-- Name: team_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.team_members (
    id bigint NOT NULL,
    contest_id bigint NOT NULL,
    team_id bigint NOT NULL,
    user_id bigint NOT NULL,
    joined_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: team_members_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.team_members_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: team_members_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.team_members_id_seq OWNED BY public.team_members.id;


--
-- Name: teams; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.teams (
    id bigint NOT NULL,
    contest_id bigint NOT NULL,
    name character varying(100) NOT NULL,
    captain_id bigint NOT NULL,
    invite_code character varying(6) NOT NULL,
    max_members integer DEFAULT 4 NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp without time zone,
    total_score integer DEFAULT 0 NOT NULL,
    last_solve_at timestamp without time zone
);


--
-- Name: teams_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.teams_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: teams_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.teams_id_seq OWNED BY public.teams.id;


--
-- Name: user_roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_roles (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    role_id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: user_roles_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_roles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_roles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_roles_id_seq OWNED BY public.user_roles.id;


--
-- Name: user_scores; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_scores (
    user_id bigint NOT NULL,
    total_score integer DEFAULT 0 NOT NULL,
    solved_count integer DEFAULT 0 NOT NULL,
    rank integer DEFAULT 0 NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    username character varying(64) NOT NULL,
    password_hash character varying(255) NOT NULL,
    email character varying(255) DEFAULT NULL::character varying,
    role character varying(32) DEFAULT 'student'::character varying NOT NULL,
    class_name character varying(128) DEFAULT NULL::character varying,
    status character varying(16) DEFAULT 'active'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    student_no character varying(64) DEFAULT NULL::character varying,
    teacher_no character varying(64) DEFAULT NULL::character varying,
    name character varying(64) DEFAULT NULL::character varying,
    failed_login_attempts integer DEFAULT 0 NOT NULL,
    last_failed_login_at timestamp with time zone,
    locked_until timestamp with time zone
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: audit_logs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs ALTER COLUMN id SET DEFAULT nextval('public.audit_logs_id_seq'::regclass);


--
-- Name: awd_attack_logs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_attack_logs ALTER COLUMN id SET DEFAULT nextval('public.awd_attack_logs_id_seq'::regclass);


--
-- Name: awd_challenges id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_challenges ALTER COLUMN id SET DEFAULT nextval('public.awd_challenges_id_seq'::regclass);


--
-- Name: awd_rounds id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_rounds ALTER COLUMN id SET DEFAULT nextval('public.awd_rounds_id_seq'::regclass);


--
-- Name: awd_service_templates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_service_templates ALTER COLUMN id SET DEFAULT nextval('public.awd_service_templates_id_seq'::regclass);


--
-- Name: awd_team_services id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_team_services ALTER COLUMN id SET DEFAULT nextval('public.awd_team_services_id_seq'::regclass);


--
-- Name: awd_traffic_events id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_traffic_events ALTER COLUMN id SET DEFAULT nextval('public.awd_traffic_events_id_seq'::regclass);


--
-- Name: challenge_hints id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_hints ALTER COLUMN id SET DEFAULT nextval('public.challenge_hints_id_seq'::regclass);


--
-- Name: challenge_package_revisions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_package_revisions ALTER COLUMN id SET DEFAULT nextval('public.challenge_package_revisions_id_seq'::regclass);


--
-- Name: challenge_publish_check_jobs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_publish_check_jobs ALTER COLUMN id SET DEFAULT nextval('public.challenge_publish_check_jobs_id_seq'::regclass);


--
-- Name: challenge_tags id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_tags ALTER COLUMN id SET DEFAULT nextval('public.challenge_tags_id_seq'::regclass);


--
-- Name: challenge_topologies id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_topologies ALTER COLUMN id SET DEFAULT nextval('public.challenge_topologies_id_seq'::regclass);


--
-- Name: challenge_writeups id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_writeups ALTER COLUMN id SET DEFAULT nextval('public.challenge_writeups_id_seq'::regclass);


--
-- Name: challenges id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenges ALTER COLUMN id SET DEFAULT nextval('public.challenges_id_seq'::regclass);


--
-- Name: contest_announcements id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contest_announcements ALTER COLUMN id SET DEFAULT nextval('public.contest_announcements_id_seq'::regclass);


--
-- Name: contest_awd_services id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contest_awd_services ALTER COLUMN id SET DEFAULT nextval('public.contest_awd_services_id_seq'::regclass);


--
-- Name: contest_challenges id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contest_challenges ALTER COLUMN id SET DEFAULT nextval('public.contest_challenges_id_seq'::regclass);


--
-- Name: contest_registrations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contest_registrations ALTER COLUMN id SET DEFAULT nextval('public.contest_registrations_id_seq'::regclass);


--
-- Name: contests id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contests ALTER COLUMN id SET DEFAULT nextval('public.contests_id_seq'::regclass);


--
-- Name: environment_templates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.environment_templates ALTER COLUMN id SET DEFAULT nextval('public.environment_templates_id_seq'::regclass);


--
-- Name: images id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.images ALTER COLUMN id SET DEFAULT nextval('public.images_id_seq'::regclass);


--
-- Name: instances id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.instances ALTER COLUMN id SET DEFAULT nextval('public.instances_id_seq'::regclass);


--
-- Name: notification_batches id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_batches ALTER COLUMN id SET DEFAULT nextval('public.notification_batches_id_seq'::regclass);


--
-- Name: notifications id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ALTER COLUMN id SET DEFAULT nextval('public.notifications_id_seq'::regclass);


--
-- Name: reports id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reports ALTER COLUMN id SET DEFAULT nextval('public.reports_id_seq'::regclass);


--
-- Name: roles id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles ALTER COLUMN id SET DEFAULT nextval('public.roles_id_seq'::regclass);


--
-- Name: skill_profiles id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.skill_profiles ALTER COLUMN id SET DEFAULT nextval('public.skill_profiles_id_seq'::regclass);


--
-- Name: submission_writeups id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.submission_writeups ALTER COLUMN id SET DEFAULT nextval('public.submission_writeups_id_seq'::regclass);


--
-- Name: submissions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.submissions ALTER COLUMN id SET DEFAULT nextval('public.submissions_id_seq'::regclass);


--
-- Name: tags id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tags ALTER COLUMN id SET DEFAULT nextval('public.tags_id_seq'::regclass);


--
-- Name: team_members id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_members ALTER COLUMN id SET DEFAULT nextval('public.team_members_id_seq'::regclass);


--
-- Name: teams id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams ALTER COLUMN id SET DEFAULT nextval('public.teams_id_seq'::regclass);


--
-- Name: user_roles id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_roles ALTER COLUMN id SET DEFAULT nextval('public.user_roles_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: awd_attack_logs awd_attack_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_attack_logs
    ADD CONSTRAINT awd_attack_logs_pkey PRIMARY KEY (id);


--
-- Name: awd_challenges awd_challenges_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_challenges
    ADD CONSTRAINT awd_challenges_pkey PRIMARY KEY (id);


--
-- Name: awd_rounds awd_rounds_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_rounds
    ADD CONSTRAINT awd_rounds_pkey PRIMARY KEY (id);


--
-- Name: awd_service_templates awd_service_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_service_templates
    ADD CONSTRAINT awd_service_templates_pkey PRIMARY KEY (id);


--
-- Name: awd_team_services awd_team_services_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_team_services
    ADD CONSTRAINT awd_team_services_pkey PRIMARY KEY (id);


--
-- Name: awd_traffic_events awd_traffic_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_traffic_events
    ADD CONSTRAINT awd_traffic_events_pkey PRIMARY KEY (id);


--
-- Name: challenge_hints challenge_hints_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_hints
    ADD CONSTRAINT challenge_hints_pkey PRIMARY KEY (id);


--
-- Name: challenge_package_revisions challenge_package_revisions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_package_revisions
    ADD CONSTRAINT challenge_package_revisions_pkey PRIMARY KEY (id);


--
-- Name: challenge_publish_check_jobs challenge_publish_check_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_publish_check_jobs
    ADD CONSTRAINT challenge_publish_check_jobs_pkey PRIMARY KEY (id);


--
-- Name: challenge_tags challenge_tags_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_tags
    ADD CONSTRAINT challenge_tags_pkey PRIMARY KEY (id);


--
-- Name: challenge_topologies challenge_topologies_challenge_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_topologies
    ADD CONSTRAINT challenge_topologies_challenge_id_key UNIQUE (challenge_id);


--
-- Name: challenge_topologies challenge_topologies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_topologies
    ADD CONSTRAINT challenge_topologies_pkey PRIMARY KEY (id);


--
-- Name: challenge_writeups challenge_writeups_challenge_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_writeups
    ADD CONSTRAINT challenge_writeups_challenge_id_key UNIQUE (challenge_id);


--
-- Name: challenge_writeups challenge_writeups_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_writeups
    ADD CONSTRAINT challenge_writeups_pkey PRIMARY KEY (id);


--
-- Name: challenges challenges_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenges
    ADD CONSTRAINT challenges_pkey PRIMARY KEY (id);


--
-- Name: contest_announcements contest_announcements_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contest_announcements
    ADD CONSTRAINT contest_announcements_pkey PRIMARY KEY (id);


--
-- Name: contest_awd_services contest_awd_services_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contest_awd_services
    ADD CONSTRAINT contest_awd_services_pkey PRIMARY KEY (id);


--
-- Name: contest_challenges contest_challenges_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contest_challenges
    ADD CONSTRAINT contest_challenges_pkey PRIMARY KEY (id);


--
-- Name: contest_registrations contest_registrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contest_registrations
    ADD CONSTRAINT contest_registrations_pkey PRIMARY KEY (id);


--
-- Name: contests contests_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contests
    ADD CONSTRAINT contests_pkey PRIMARY KEY (id);


--
-- Name: environment_templates environment_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.environment_templates
    ADD CONSTRAINT environment_templates_pkey PRIMARY KEY (id);


--
-- Name: images images_name_tag_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.images
    ADD CONSTRAINT images_name_tag_key UNIQUE (name, tag);


--
-- Name: images images_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.images
    ADD CONSTRAINT images_pkey PRIMARY KEY (id);


--
-- Name: instances instances_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.instances
    ADD CONSTRAINT instances_pkey PRIMARY KEY (id);


--
-- Name: notification_batches notification_batches_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_batches
    ADD CONSTRAINT notification_batches_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: port_allocations port_allocations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.port_allocations
    ADD CONSTRAINT port_allocations_pkey PRIMARY KEY (port);


--
-- Name: reports reports_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_pkey PRIMARY KEY (id);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- Name: skill_profiles skill_profiles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.skill_profiles
    ADD CONSTRAINT skill_profiles_pkey PRIMARY KEY (id);


--
-- Name: submission_writeups submission_writeups_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.submission_writeups
    ADD CONSTRAINT submission_writeups_pkey PRIMARY KEY (id);


--
-- Name: submissions submissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.submissions
    ADD CONSTRAINT submissions_pkey PRIMARY KEY (id);


--
-- Name: tags tags_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_pkey PRIMARY KEY (id);


--
-- Name: team_members team_members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_members
    ADD CONSTRAINT team_members_pkey PRIMARY KEY (id);


--
-- Name: teams teams_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT teams_pkey PRIMARY KEY (id);


--
-- Name: team_members uk_team_members_contest_user; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_members
    ADD CONSTRAINT uk_team_members_contest_user UNIQUE (contest_id, user_id);


--
-- Name: team_members uk_team_members_team_user; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_members
    ADD CONSTRAINT uk_team_members_team_user UNIQUE (team_id, user_id);


--
-- Name: user_roles user_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_pkey PRIMARY KEY (id);


--
-- Name: user_scores user_scores_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_scores
    ADD CONSTRAINT user_scores_pkey PRIMARY KEY (user_id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_audit_logs_action; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_action ON public.audit_logs USING btree (action, created_at DESC);


--
-- Name: idx_audit_logs_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_created ON public.audit_logs USING btree (created_at);


--
-- Name: idx_audit_logs_resource; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_resource ON public.audit_logs USING btree (resource_type, resource_id, created_at DESC);


--
-- Name: idx_audit_logs_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_user ON public.audit_logs USING btree (user_id, created_at DESC);


--
-- Name: idx_awd_attack_round; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_awd_attack_round ON public.awd_attack_logs USING btree (round_id, attacker_team_id);


--
-- Name: idx_awd_attack_round_service_success; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_awd_attack_round_service_success ON public.awd_attack_logs USING btree (round_id, attacker_team_id, victim_team_id, service_id, is_success);


--
-- Name: idx_awd_attack_source; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_awd_attack_source ON public.awd_attack_logs USING btree (round_id, source);


--
-- Name: idx_awd_attack_submitter_success; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_awd_attack_submitter_success ON public.awd_attack_logs USING btree (submitted_by_user_id, is_success, score_gained);


--
-- Name: idx_awd_attack_success; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_awd_attack_success ON public.awd_attack_logs USING btree (round_id, is_success) WHERE (is_success = true);


--
-- Name: idx_awd_attack_victim; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_awd_attack_victim ON public.awd_attack_logs USING btree (round_id, victim_team_id);


--
-- Name: idx_awd_challenges_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_awd_challenges_status ON public.awd_challenges USING btree (status, service_type);


--
-- Name: idx_awd_rounds_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_awd_rounds_status ON public.awd_rounds USING btree (contest_id, status);


--
-- Name: idx_awd_service_templates_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_awd_service_templates_status ON public.awd_service_templates USING btree (status, service_type);


--
-- Name: idx_awd_traffic_attacker; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_awd_traffic_attacker ON public.awd_traffic_events USING btree (round_id, attacker_team_id);


--
-- Name: idx_awd_traffic_round_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_awd_traffic_round_created ON public.awd_traffic_events USING btree (round_id, created_at DESC, id DESC);


--
-- Name: idx_awd_traffic_round_summary; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_awd_traffic_round_summary ON public.awd_traffic_events USING btree (round_id, method, path, status_code);


--
-- Name: idx_awd_traffic_service; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_awd_traffic_service ON public.awd_traffic_events USING btree (service_id);


--
-- Name: idx_awd_traffic_victim; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_awd_traffic_victim ON public.awd_traffic_events USING btree (round_id, victim_team_id);


--
-- Name: idx_awd_ts_round_team_service; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_awd_ts_round_team_service ON public.awd_team_services USING btree (round_id, team_id, service_id);


--
-- Name: idx_awd_ts_team; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_awd_ts_team ON public.awd_team_services USING btree (team_id, round_id);


--
-- Name: idx_challenge_hints_challenge_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_challenge_hints_challenge_id ON public.challenge_hints USING btree (challenge_id);


--
-- Name: idx_challenge_package_revisions_challenge_revision; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_challenge_package_revisions_challenge_revision ON public.challenge_package_revisions USING btree (challenge_id, revision_no);


--
-- Name: idx_challenge_tags_tag_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_challenge_tags_tag_id ON public.challenge_tags USING btree (tag_id);


--
-- Name: idx_challenge_topologies_last_export_revision_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_challenge_topologies_last_export_revision_id ON public.challenge_topologies USING btree (last_export_revision_id);


--
-- Name: idx_challenge_topologies_package_revision_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_challenge_topologies_package_revision_id ON public.challenge_topologies USING btree (package_revision_id);


--
-- Name: idx_challenge_topologies_template_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_challenge_topologies_template_id ON public.challenge_topologies USING btree (template_id);


--
-- Name: idx_challenge_writeups_recommended; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_challenge_writeups_recommended ON public.challenge_writeups USING btree (is_recommended, recommended_at DESC);


--
-- Name: idx_challenges_category; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_challenges_category ON public.challenges USING btree (category) WHERE (deleted_at IS NULL);


--
-- Name: idx_challenges_created_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_challenges_created_by ON public.challenges USING btree (created_by);


--
-- Name: idx_challenges_difficulty; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_challenges_difficulty ON public.challenges USING btree (difficulty) WHERE (deleted_at IS NULL);


--
-- Name: idx_challenges_flag_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_challenges_flag_type ON public.challenges USING btree (flag_type);


--
-- Name: idx_challenges_image_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_challenges_image_id ON public.challenges USING btree (image_id);


--
-- Name: idx_challenges_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_challenges_status ON public.challenges USING btree (status) WHERE (deleted_at IS NULL);


--
-- Name: idx_contest_announcements_contest_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_contest_announcements_contest_created ON public.contest_announcements USING btree (contest_id, created_at DESC);


--
-- Name: idx_contest_awd_services_awd_challenge; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_contest_awd_services_awd_challenge ON public.contest_awd_services USING btree (awd_challenge_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_contest_awd_services_contest_order; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_contest_awd_services_contest_order ON public.contest_awd_services USING btree (contest_id, "order", id) WHERE (deleted_at IS NULL);


--
-- Name: idx_contest_challenges_active_order; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_contest_challenges_active_order ON public.contest_challenges USING btree (contest_id, "order") WHERE (deleted_at IS NULL);


--
-- Name: idx_contest_challenges_active_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_contest_challenges_active_unique ON public.contest_challenges USING btree (contest_id, challenge_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_contest_challenges_contest_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_contest_challenges_contest_id ON public.contest_challenges USING btree (contest_id);


--
-- Name: idx_contest_reg_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_contest_reg_status ON public.contest_registrations USING btree (contest_id, status);


--
-- Name: idx_contests_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_contests_deleted_at ON public.contests USING btree (deleted_at);


--
-- Name: idx_contests_start_time; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_contests_start_time ON public.contests USING btree (start_time);


--
-- Name: idx_contests_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_contests_status ON public.contests USING btree (status);


--
-- Name: idx_cp_jobs_challenge_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_cp_jobs_challenge_active ON public.challenge_publish_check_jobs USING btree (challenge_id) WHERE ((status)::text = ANY (ARRAY[('pending'::character varying)::text, ('running'::character varying)::text]));


--
-- Name: idx_cp_jobs_challenge_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_cp_jobs_challenge_created ON public.challenge_publish_check_jobs USING btree (challenge_id, created_at DESC);


--
-- Name: idx_cp_jobs_status_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_cp_jobs_status_created ON public.challenge_publish_check_jobs USING btree (status, created_at);


--
-- Name: idx_environment_templates_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_environment_templates_name ON public.environment_templates USING btree (name);


--
-- Name: idx_images_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_images_deleted_at ON public.images USING btree (deleted_at);


--
-- Name: idx_images_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_images_name ON public.images USING btree (name);


--
-- Name: idx_images_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_images_status ON public.images USING btree (status);


--
-- Name: idx_instances_challenge_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_instances_challenge_id ON public.instances USING btree (challenge_id);


--
-- Name: idx_instances_contest_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_instances_contest_id ON public.instances USING btree (contest_id) WHERE (contest_id IS NOT NULL);


--
-- Name: idx_instances_destroyed_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_instances_destroyed_at ON public.instances USING btree (destroyed_at);


--
-- Name: idx_instances_expires_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_instances_expires_at ON public.instances USING btree (expires_at);


--
-- Name: idx_instances_service_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_instances_service_id ON public.instances USING btree (service_id) WHERE (service_id IS NOT NULL);


--
-- Name: idx_instances_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_instances_status ON public.instances USING btree (status);


--
-- Name: idx_instances_team_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_instances_team_id ON public.instances USING btree (team_id) WHERE (team_id IS NOT NULL);


--
-- Name: idx_instances_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_instances_user_id ON public.instances USING btree (user_id);


--
-- Name: idx_name_tag; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_name_tag ON public.images USING btree (name, tag);


--
-- Name: idx_notification_batches_created_by_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notification_batches_created_by_created_at ON public.notification_batches USING btree (created_by, created_at DESC);


--
-- Name: idx_notifications_batch_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_batch_id ON public.notifications USING btree (batch_id);


--
-- Name: idx_notifications_user_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_user_created_at ON public.notifications USING btree (user_id, created_at DESC);


--
-- Name: idx_notifications_user_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_user_type ON public.notifications USING btree (user_id, type, created_at DESC);


--
-- Name: idx_notifications_user_unread; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_user_unread ON public.notifications USING btree (user_id, is_read, created_at DESC);


--
-- Name: idx_port_allocations_instance_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_port_allocations_instance_id ON public.port_allocations USING btree (instance_id);


--
-- Name: idx_reports_class_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_reports_class_status ON public.reports USING btree (class_name, status, created_at DESC);


--
-- Name: idx_reports_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_reports_created_at ON public.reports USING btree (created_at DESC);


--
-- Name: idx_reports_user_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_reports_user_status ON public.reports USING btree (user_id, status, created_at DESC);


--
-- Name: idx_skill_profiles_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_skill_profiles_updated_at ON public.skill_profiles USING btree (updated_at);


--
-- Name: idx_skill_profiles_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_skill_profiles_user_id ON public.skill_profiles USING btree (user_id);


--
-- Name: idx_submission_writeups_challenge; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_submission_writeups_challenge ON public.submission_writeups USING btree (challenge_id, updated_at DESC);


--
-- Name: idx_submission_writeups_recommended; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_submission_writeups_recommended ON public.submission_writeups USING btree (is_recommended, recommended_at DESC);


--
-- Name: idx_submission_writeups_user_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_submission_writeups_user_updated_at ON public.submission_writeups USING btree (user_id, updated_at DESC);


--
-- Name: idx_submission_writeups_visibility_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_submission_writeups_visibility_status ON public.submission_writeups USING btree (visibility_status, updated_at DESC);


--
-- Name: idx_submissions_challenge_correct; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_submissions_challenge_correct ON public.submissions USING btree (challenge_id, is_correct);


--
-- Name: idx_submissions_contest_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_submissions_contest_id ON public.submissions USING btree (contest_id) WHERE (contest_id IS NOT NULL);


--
-- Name: idx_submissions_review_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_submissions_review_status ON public.submissions USING btree (review_status, updated_at DESC);


--
-- Name: idx_submissions_submitted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_submissions_submitted_at ON public.submissions USING btree (submitted_at);


--
-- Name: idx_submissions_user_challenge; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_submissions_user_challenge ON public.submissions USING btree (user_id, challenge_id);


--
-- Name: idx_submissions_user_challenge_correct; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_submissions_user_challenge_correct ON public.submissions USING btree (user_id, challenge_id) WHERE (is_correct = true);


--
-- Name: idx_team_members_team_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_team_members_team_id ON public.team_members USING btree (team_id);


--
-- Name: idx_team_members_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_team_members_user_id ON public.team_members USING btree (user_id);


--
-- Name: idx_teams_captain_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_teams_captain_id ON public.teams USING btree (captain_id);


--
-- Name: idx_teams_contest_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_teams_contest_id ON public.teams USING btree (contest_id);


--
-- Name: idx_user_roles_role_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_roles_role_id ON public.user_roles USING btree (role_id);


--
-- Name: idx_user_scores_rank; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_scores_rank ON public.user_scores USING btree (rank);


--
-- Name: idx_user_scores_total_score; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_scores_total_score ON public.user_scores USING btree (total_score DESC);


--
-- Name: idx_users_locked_until; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_locked_until ON public.users USING btree (locked_until) WHERE ((deleted_at IS NULL) AND (locked_until IS NOT NULL));


--
-- Name: idx_users_role; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_role ON public.users USING btree (role) WHERE (deleted_at IS NULL);


--
-- Name: idx_users_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_status ON public.users USING btree (status) WHERE (deleted_at IS NULL);


--
-- Name: uk_awd_challenges_slug; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_awd_challenges_slug ON public.awd_challenges USING btree (slug) WHERE (deleted_at IS NULL);


--
-- Name: uk_awd_rounds; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_awd_rounds ON public.awd_rounds USING btree (contest_id, round_number);


--
-- Name: uk_awd_service_templates_slug; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_awd_service_templates_slug ON public.awd_service_templates USING btree (slug) WHERE (deleted_at IS NULL);


--
-- Name: uk_awd_team_services; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_awd_team_services ON public.awd_team_services USING btree (round_id, team_id, service_id);


--
-- Name: uk_challenge_hint_level; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_challenge_hint_level ON public.challenge_hints USING btree (challenge_id, level);


--
-- Name: uk_challenge_hints_challenge_level; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_challenge_hints_challenge_level ON public.challenge_hints USING btree (challenge_id, level);


--
-- Name: uk_challenge_tags; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_challenge_tags ON public.challenge_tags USING btree (challenge_id, tag_id);


--
-- Name: uk_contest_awd_services_contest_awd_challenge; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_contest_awd_services_contest_awd_challenge ON public.contest_awd_services USING btree (contest_id, awd_challenge_id) WHERE (deleted_at IS NULL);


--
-- Name: uk_contest_reg_user; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_contest_reg_user ON public.contest_registrations USING btree (contest_id, user_id);


--
-- Name: uk_instances_active_host_port; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_instances_active_host_port ON public.instances USING btree (host_port) WHERE ((host_port > 0) AND ((status)::text = ANY (ARRAY[('creating'::character varying)::text, ('running'::character varying)::text])));


--
-- Name: uk_instances_contest_team_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_instances_contest_team_active ON public.instances USING btree (contest_id, team_id, challenge_id) WHERE ((team_id IS NOT NULL) AND ((share_scope)::text = 'per_team'::text) AND ((status)::text = ANY (ARRAY[('creating'::character varying)::text, ('running'::character varying)::text])));


--
-- Name: uk_instances_contest_user_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_instances_contest_user_active ON public.instances USING btree (contest_id, user_id, challenge_id) WHERE ((contest_id IS NOT NULL) AND (team_id IS NULL) AND ((share_scope)::text = 'per_user'::text) AND ((status)::text = ANY (ARRAY[('creating'::character varying)::text, ('running'::character varying)::text])));


--
-- Name: uk_instances_personal_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_instances_personal_active ON public.instances USING btree (user_id, challenge_id) WHERE ((contest_id IS NULL) AND (team_id IS NULL) AND ((share_scope)::text = 'per_user'::text) AND ((status)::text = ANY (ARRAY[('creating'::character varying)::text, ('running'::character varying)::text])));


--
-- Name: uk_instances_shared_contest_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_instances_shared_contest_active ON public.instances USING btree (contest_id, challenge_id) WHERE ((contest_id IS NOT NULL) AND (team_id IS NULL) AND ((share_scope)::text = 'shared'::text) AND ((status)::text = ANY (ARRAY[('creating'::character varying)::text, ('running'::character varying)::text])));


--
-- Name: uk_instances_shared_practice_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_instances_shared_practice_active ON public.instances USING btree (challenge_id) WHERE ((contest_id IS NULL) AND (team_id IS NULL) AND ((share_scope)::text = 'shared'::text) AND ((status)::text = ANY (ARRAY[('creating'::character varying)::text, ('running'::character varying)::text])));


--
-- Name: uk_notifications_batch_user; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_notifications_batch_user ON public.notifications USING btree (batch_id, user_id) WHERE (batch_id IS NOT NULL);


--
-- Name: uk_roles_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_roles_code ON public.roles USING btree (code);


--
-- Name: uk_skill_profiles_user_dimension; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_skill_profiles_user_dimension ON public.skill_profiles USING btree (user_id, dimension);


--
-- Name: uk_submission_writeups_user_challenge; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_submission_writeups_user_challenge ON public.submission_writeups USING btree (user_id, challenge_id);


--
-- Name: uk_submissions_contest_team_challenge_correct; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_submissions_contest_team_challenge_correct ON public.submissions USING btree (contest_id, team_id, challenge_id) WHERE ((is_correct = true) AND (contest_id IS NOT NULL) AND (team_id IS NOT NULL));


--
-- Name: uk_submissions_contest_user_challenge_correct; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_submissions_contest_user_challenge_correct ON public.submissions USING btree (contest_id, user_id, challenge_id) WHERE ((is_correct = true) AND (contest_id IS NOT NULL) AND (team_id IS NULL));


--
-- Name: uk_tags_name_type; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_tags_name_type ON public.tags USING btree (name, type);


--
-- Name: uk_teams_contest_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_teams_contest_name ON public.teams USING btree (contest_id, name) WHERE (deleted_at IS NULL);


--
-- Name: uk_teams_invite_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_teams_invite_code ON public.teams USING btree (invite_code);


--
-- Name: uk_user_roles; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_user_roles ON public.user_roles USING btree (user_id, role_id);


--
-- Name: uk_users_email; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_users_email ON public.users USING btree (email) WHERE ((deleted_at IS NULL) AND (email IS NOT NULL) AND (btrim((email)::text) <> ''::text));


--
-- Name: uk_users_student_no; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_users_student_no ON public.users USING btree (student_no) WHERE ((deleted_at IS NULL) AND (student_no IS NOT NULL) AND ((student_no)::text <> ''::text));


--
-- Name: uk_users_teacher_no; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_users_teacher_no ON public.users USING btree (teacher_no) WHERE ((deleted_at IS NULL) AND (teacher_no IS NOT NULL) AND ((teacher_no)::text <> ''::text));


--
-- Name: uk_users_username; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_users_username ON public.users USING btree (username) WHERE (deleted_at IS NULL);


--
-- Name: uq_challenges_package_slug; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uq_challenges_package_slug ON public.challenges USING btree (package_slug);


--
-- Name: awd_attack_logs awd_attack_logs_attacker_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_attack_logs
    ADD CONSTRAINT awd_attack_logs_attacker_team_id_fkey FOREIGN KEY (attacker_team_id) REFERENCES public.teams(id) ON DELETE CASCADE;


--
-- Name: awd_attack_logs awd_attack_logs_awd_challenge_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_attack_logs
    ADD CONSTRAINT awd_attack_logs_awd_challenge_id_fkey FOREIGN KEY (awd_challenge_id) REFERENCES public.awd_challenges(id) ON DELETE RESTRICT;


--
-- Name: awd_attack_logs awd_attack_logs_round_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_attack_logs
    ADD CONSTRAINT awd_attack_logs_round_id_fkey FOREIGN KEY (round_id) REFERENCES public.awd_rounds(id) ON DELETE CASCADE;


--
-- Name: awd_attack_logs awd_attack_logs_victim_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_attack_logs
    ADD CONSTRAINT awd_attack_logs_victim_team_id_fkey FOREIGN KEY (victim_team_id) REFERENCES public.teams(id) ON DELETE CASCADE;


--
-- Name: awd_rounds awd_rounds_contest_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_rounds
    ADD CONSTRAINT awd_rounds_contest_id_fkey FOREIGN KEY (contest_id) REFERENCES public.contests(id) ON DELETE CASCADE;


--
-- Name: awd_team_services awd_team_services_awd_challenge_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_team_services
    ADD CONSTRAINT awd_team_services_awd_challenge_id_fkey FOREIGN KEY (awd_challenge_id) REFERENCES public.awd_challenges(id) ON DELETE RESTRICT;


--
-- Name: awd_team_services awd_team_services_round_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_team_services
    ADD CONSTRAINT awd_team_services_round_id_fkey FOREIGN KEY (round_id) REFERENCES public.awd_rounds(id) ON DELETE CASCADE;


--
-- Name: awd_team_services awd_team_services_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_team_services
    ADD CONSTRAINT awd_team_services_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(id) ON DELETE CASCADE;


--
-- Name: awd_traffic_events awd_traffic_events_attacker_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_traffic_events
    ADD CONSTRAINT awd_traffic_events_attacker_team_id_fkey FOREIGN KEY (attacker_team_id) REFERENCES public.teams(id) ON DELETE CASCADE;


--
-- Name: awd_traffic_events awd_traffic_events_awd_challenge_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_traffic_events
    ADD CONSTRAINT awd_traffic_events_awd_challenge_id_fkey FOREIGN KEY (awd_challenge_id) REFERENCES public.awd_challenges(id) ON DELETE RESTRICT;


--
-- Name: awd_traffic_events awd_traffic_events_contest_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_traffic_events
    ADD CONSTRAINT awd_traffic_events_contest_id_fkey FOREIGN KEY (contest_id) REFERENCES public.contests(id) ON DELETE CASCADE;


--
-- Name: awd_traffic_events awd_traffic_events_round_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_traffic_events
    ADD CONSTRAINT awd_traffic_events_round_id_fkey FOREIGN KEY (round_id) REFERENCES public.awd_rounds(id) ON DELETE CASCADE;


--
-- Name: awd_traffic_events awd_traffic_events_victim_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awd_traffic_events
    ADD CONSTRAINT awd_traffic_events_victim_team_id_fkey FOREIGN KEY (victim_team_id) REFERENCES public.teams(id) ON DELETE CASCADE;


--
-- Name: challenge_package_revisions challenge_package_revisions_challenge_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_package_revisions
    ADD CONSTRAINT challenge_package_revisions_challenge_id_fkey FOREIGN KEY (challenge_id) REFERENCES public.challenges(id) ON DELETE CASCADE;


--
-- Name: challenge_package_revisions challenge_package_revisions_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_package_revisions
    ADD CONSTRAINT challenge_package_revisions_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: challenge_package_revisions challenge_package_revisions_parent_revision_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_package_revisions
    ADD CONSTRAINT challenge_package_revisions_parent_revision_id_fkey FOREIGN KEY (parent_revision_id) REFERENCES public.challenge_package_revisions(id) ON DELETE SET NULL;


--
-- Name: challenge_publish_check_jobs challenge_publish_check_jobs_challenge_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_publish_check_jobs
    ADD CONSTRAINT challenge_publish_check_jobs_challenge_id_fkey FOREIGN KEY (challenge_id) REFERENCES public.challenges(id) ON DELETE CASCADE;


--
-- Name: challenge_publish_check_jobs challenge_publish_check_jobs_requested_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_publish_check_jobs
    ADD CONSTRAINT challenge_publish_check_jobs_requested_by_fkey FOREIGN KEY (requested_by) REFERENCES public.users(id) ON DELETE RESTRICT;


--
-- Name: challenge_topologies challenge_topologies_challenge_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_topologies
    ADD CONSTRAINT challenge_topologies_challenge_id_fkey FOREIGN KEY (challenge_id) REFERENCES public.challenges(id) ON DELETE CASCADE;


--
-- Name: challenge_topologies challenge_topologies_last_export_revision_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_topologies
    ADD CONSTRAINT challenge_topologies_last_export_revision_id_fkey FOREIGN KEY (last_export_revision_id) REFERENCES public.challenge_package_revisions(id) ON DELETE SET NULL;


--
-- Name: challenge_topologies challenge_topologies_package_revision_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_topologies
    ADD CONSTRAINT challenge_topologies_package_revision_id_fkey FOREIGN KEY (package_revision_id) REFERENCES public.challenge_package_revisions(id) ON DELETE SET NULL;


--
-- Name: challenge_topologies challenge_topologies_template_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_topologies
    ADD CONSTRAINT challenge_topologies_template_id_fkey FOREIGN KEY (template_id) REFERENCES public.environment_templates(id) ON DELETE SET NULL;


--
-- Name: challenge_writeups challenge_writeups_challenge_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_writeups
    ADD CONSTRAINT challenge_writeups_challenge_id_fkey FOREIGN KEY (challenge_id) REFERENCES public.challenges(id) ON DELETE CASCADE;


--
-- Name: challenge_writeups challenge_writeups_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_writeups
    ADD CONSTRAINT challenge_writeups_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: challenge_writeups challenge_writeups_recommended_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenge_writeups
    ADD CONSTRAINT challenge_writeups_recommended_by_fkey FOREIGN KEY (recommended_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: challenges challenges_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.challenges
    ADD CONSTRAINT challenges_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: contest_announcements contest_announcements_contest_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contest_announcements
    ADD CONSTRAINT contest_announcements_contest_id_fkey FOREIGN KEY (contest_id) REFERENCES public.contests(id) ON DELETE CASCADE;


--
-- Name: contest_announcements contest_announcements_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contest_announcements
    ADD CONSTRAINT contest_announcements_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: contest_challenges fk_contest_challenges_challenge; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contest_challenges
    ADD CONSTRAINT fk_contest_challenges_challenge FOREIGN KEY (challenge_id) REFERENCES public.challenges(id) ON DELETE CASCADE;


--
-- Name: contest_challenges fk_contest_challenges_contest; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contest_challenges
    ADD CONSTRAINT fk_contest_challenges_contest FOREIGN KEY (contest_id) REFERENCES public.contests(id) ON DELETE CASCADE;


--
-- Name: team_members fk_team_members_team; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_members
    ADD CONSTRAINT fk_team_members_team FOREIGN KEY (team_id) REFERENCES public.teams(id) ON DELETE CASCADE;


--
-- Name: teams fk_teams_contest; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT fk_teams_contest FOREIGN KEY (contest_id) REFERENCES public.contests(id) ON DELETE CASCADE;


--
-- Name: notification_batches notification_batches_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_batches
    ADD CONSTRAINT notification_batches_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE RESTRICT;


--
-- Name: notifications notifications_batch_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_batch_id_fkey FOREIGN KEY (batch_id) REFERENCES public.notification_batches(id) ON DELETE SET NULL;


--
-- Name: notifications notifications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: port_allocations port_allocations_instance_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.port_allocations
    ADD CONSTRAINT port_allocations_instance_id_fkey FOREIGN KEY (instance_id) REFERENCES public.instances(id) ON DELETE CASCADE;


--
-- Name: skill_profiles skill_profiles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.skill_profiles
    ADD CONSTRAINT skill_profiles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: submission_writeups submission_writeups_challenge_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.submission_writeups
    ADD CONSTRAINT submission_writeups_challenge_id_fkey FOREIGN KEY (challenge_id) REFERENCES public.challenges(id) ON DELETE CASCADE;


--
-- Name: submission_writeups submission_writeups_contest_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.submission_writeups
    ADD CONSTRAINT submission_writeups_contest_id_fkey FOREIGN KEY (contest_id) REFERENCES public.contests(id) ON DELETE SET NULL;


--
-- Name: submission_writeups submission_writeups_recommended_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.submission_writeups
    ADD CONSTRAINT submission_writeups_recommended_by_fkey FOREIGN KEY (recommended_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: submission_writeups submission_writeups_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.submission_writeups
    ADD CONSTRAINT submission_writeups_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: submissions submissions_reviewed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.submissions
    ADD CONSTRAINT submissions_reviewed_by_fkey FOREIGN KEY (reviewed_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: user_roles user_roles_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE RESTRICT;


--
-- Name: user_roles user_roles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_scores user_scores_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_scores
    ADD CONSTRAINT user_scores_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


INSERT INTO public.roles (code, name, description)
VALUES
    ('student', 'Student', '学员'),
    ('teacher', 'Teacher', '教师'),
    ('admin', 'Admin', '管理员')
ON CONFLICT (code) DO NOTHING;

INSERT INTO public.users (
    username,
    password_hash,
    email,
    role,
    class_name,
    status,
    student_no,
    teacher_no,
    name
)
VALUES
    ('admin', '$2a$10$BZtdXbWm6DBx6NWilSbaqOTecF3c8xnhnXmYBZicjjpiXHwNCD9fu', 'admin@example.com', 'admin', NULL, 'active', NULL, NULL, 'Platform Admin'),
    ('teacher', '$2a$10$BZtdXbWm6DBx6NWilSbaqOTecF3c8xnhnXmYBZicjjpiXHwNCD9fu', 'teacher@example.com', 'teacher', 'CTF-1', 'active', NULL, 'T20260001', 'Demo Teacher'),
    ('student', '$2a$10$BZtdXbWm6DBx6NWilSbaqOTecF3c8xnhnXmYBZicjjpiXHwNCD9fu', 'student@example.com', 'student', 'CTF-1', 'active', '20260001', NULL, 'Demo Student'),
    ('student2', '$2a$10$BZtdXbWm6DBx6NWilSbaqOTecF3c8xnhnXmYBZicjjpiXHwNCD9fu', 'student2@example.com', 'student', 'CTF-1', 'active', '20260002', NULL, 'Demo Student 2')
ON CONFLICT (username) WHERE (deleted_at IS NULL) DO UPDATE
SET
    password_hash = EXCLUDED.password_hash,
    email = EXCLUDED.email,
    role = EXCLUDED.role,
    class_name = EXCLUDED.class_name,
    status = EXCLUDED.status,
    student_no = EXCLUDED.student_no,
    teacher_no = EXCLUDED.teacher_no,
    name = EXCLUDED.name,
    updated_at = now();

DELETE FROM public.user_roles
WHERE user_id IN (
    SELECT id
    FROM public.users
    WHERE username IN ('admin', 'teacher', 'student', 'student2')
);

INSERT INTO public.user_roles (user_id, role_id)
SELECT
    u.id,
    r.id
FROM public.users AS u
JOIN public.roles AS r
    ON r.code = u.role
WHERE u.username IN ('admin', 'teacher', 'student', 'student2')
ON CONFLICT (user_id, role_id) DO NOTHING;


--
-- PostgreSQL database dump complete
--
