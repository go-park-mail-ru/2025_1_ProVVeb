--
-- PostgreSQL database dump
--

-- Dumped from database version 16.9 (Debian 16.9-1.pgdg120+1)
-- Dumped by pg_dump version 16.9 (Debian 16.9-1.pgdg120+1)

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
-- Name: pg_stat_statements; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_stat_statements WITH SCHEMA public;


--
-- Name: EXTENSION pg_stat_statements; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pg_stat_statements IS 'track planning and execution statistics of all SQL statements executed';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: admins; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.admins (
    admin_id bigint NOT NULL,
    user_id bigint NOT NULL,
    role text NOT NULL,
    CONSTRAINT admins_role_check CHECK ((length(role) <= 40))
);


--
-- Name: admins_admin_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.admins ALTER COLUMN admin_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.admins_admin_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: blacklist; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.blacklist (
    block_id bigint NOT NULL,
    user_id bigint NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: blacklist_block_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.blacklist ALTER COLUMN block_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.blacklist_block_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: chats; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.chats (
    chat_id bigint NOT NULL,
    first_profile_id bigint NOT NULL,
    second_profile_id bigint NOT NULL,
    last_message text NOT NULL,
    last_sender bigint NOT NULL,
    CONSTRAINT chats_last_message_check CHECK ((length(last_message) <= 400))
);


--
-- Name: chats_chat_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.chats ALTER COLUMN chat_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.chats_chat_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: complaint_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.complaint_types (
    comp_type bigint NOT NULL,
    type_description text NOT NULL,
    CONSTRAINT complaint_types_type_description_check CHECK ((length(type_description) <= 255))
);


--
-- Name: complaint_types_comp_type_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.complaint_types ALTER COLUMN comp_type ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.complaint_types_comp_type_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: complaints; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.complaints (
    complaint_id bigint NOT NULL,
    complaint_by bigint NOT NULL,
    complaint_on bigint NOT NULL,
    complaint_type bigint NOT NULL,
    complaint_text text NOT NULL,
    status integer NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    closed_at timestamp without time zone
);


--
-- Name: complaints_complaint_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.complaints ALTER COLUMN complaint_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.complaints_complaint_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: interests; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.interests (
    interest_id bigint NOT NULL,
    description text NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: interests_interest_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.interests ALTER COLUMN interest_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.interests_interest_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: likes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.likes (
    like_id bigint NOT NULL,
    profile_id bigint NOT NULL,
    liked_profile_id bigint NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    status integer NOT NULL
);


--
-- Name: likes_like_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.likes ALTER COLUMN like_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.likes_like_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: locations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.locations (
    location_id bigint NOT NULL,
    country text NOT NULL,
    city text NOT NULL,
    district text NOT NULL,
    CONSTRAINT locations_city_check CHECK ((length(city) <= 255)),
    CONSTRAINT locations_country_check CHECK ((length(country) <= 255)),
    CONSTRAINT locations_district_check CHECK ((length(district) <= 255))
);


--
-- Name: locations_location_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.locations ALTER COLUMN location_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.locations_location_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: matches; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.matches (
    match_id bigint NOT NULL,
    profile_id bigint NOT NULL,
    matched_profile_id bigint NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: matches_match_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.matches ALTER COLUMN match_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.matches_match_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.messages (
    message_id bigint NOT NULL,
    chat_id bigint NOT NULL,
    user_id bigint NOT NULL,
    content text NOT NULL,
    status integer NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: messages_message_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.messages ALTER COLUMN message_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.messages_message_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: notification_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notification_types (
    notif_type bigint NOT NULL,
    type_description text NOT NULL,
    CONSTRAINT notification_types_type_description_check CHECK ((length(type_description) <= 255))
);


--
-- Name: notification_types_notif_type_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.notification_types ALTER COLUMN notif_type ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.notification_types_notif_type_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    notification_id bigint NOT NULL,
    user_id bigint NOT NULL,
    notification_type bigint NOT NULL,
    content text NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    read_at timestamp without time zone
);


--
-- Name: notifications_notification_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.notifications ALTER COLUMN notification_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.notifications_notification_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: preferences; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.preferences (
    preference_id bigint NOT NULL,
    preference_type integer NOT NULL,
    preference_description text NOT NULL,
    preference_value text NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT preferences_preference_description_check CHECK ((length(preference_description) <= 255)),
    CONSTRAINT preferences_preference_value_check CHECK ((length(preference_value) <= 255))
);


--
-- Name: preferences_preference_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.preferences ALTER COLUMN preference_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.preferences_preference_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: profile_interests; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.profile_interests (
    profile_id bigint NOT NULL,
    interest_id bigint NOT NULL
);


--
-- Name: profile_preferences; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.profile_preferences (
    profile_id bigint NOT NULL,
    preference_id bigint NOT NULL
);


--
-- Name: profile_ratings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.profile_ratings (
    rating_id bigint NOT NULL,
    profile_id bigint NOT NULL,
    rated_profile_id bigint NOT NULL,
    rating_score integer NOT NULL,
    comment text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: profile_ratings_rating_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.profile_ratings ALTER COLUMN rating_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.profile_ratings_rating_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: profiles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.profiles (
    profile_id bigint NOT NULL,
    firstname text NOT NULL,
    lastname text NOT NULL,
    is_male boolean NOT NULL,
    birthday date NOT NULL,
    height integer,
    description text,
    location_id bigint,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT profiles_firstname_check CHECK ((length(firstname) <= 255)),
    CONSTRAINT profiles_height_check CHECK (((height >= 50) AND (height <= 280))),
    CONSTRAINT profiles_lastname_check CHECK ((length(lastname) <= 255))
);


--
-- Name: profiles_profile_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.profiles ALTER COLUMN profile_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.profiles_profile_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: queries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.queries (
    query_id bigint NOT NULL,
    name text NOT NULL,
    description text NOT NULL,
    min_score integer NOT NULL,
    max_score integer NOT NULL,
    is_active boolean NOT NULL,
    CONSTRAINT queries_name_check CHECK ((length(name) <= 255))
);


--
-- Name: queries_query_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.queries ALTER COLUMN query_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.queries_query_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sessions (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    token text NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    expires_at timestamp without time zone,
    CONSTRAINT sessions_token_check CHECK ((length(token) <= 255))
);


--
-- Name: sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.sessions ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: static; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.static (
    id bigint NOT NULL,
    profile_id bigint NOT NULL,
    path text DEFAULT '/default.png'::text NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT static_path_check CHECK ((length(path) <= 255))
);


--
-- Name: static_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.static ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.static_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: subscription_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.subscription_types (
    sub_type bigint NOT NULL,
    type_description text NOT NULL,
    CONSTRAINT subscription_types_type_description_check CHECK ((length(type_description) <= 255))
);


--
-- Name: subscription_types_sub_type_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.subscription_types ALTER COLUMN sub_type ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.subscription_types_sub_type_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: subscriptions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.subscriptions (
    sub_id bigint NOT NULL,
    user_id bigint NOT NULL,
    sub_type bigint NOT NULL,
    transaction_data text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    expires_at timestamp without time zone
);


--
-- Name: subscriptions_sub_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.subscriptions ALTER COLUMN sub_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.subscriptions_sub_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: user_answer; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_answer (
    answer_id bigint NOT NULL,
    query_id bigint NOT NULL,
    user_id bigint NOT NULL,
    score integer NOT NULL,
    answer text NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: user_answer_answer_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.user_answer ALTER COLUMN answer_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.user_answer_answer_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    user_id bigint NOT NULL,
    profile_id bigint,
    status integer NOT NULL,
    login text NOT NULL,
    email text NOT NULL,
    phone text,
    password text NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT users_email_check CHECK ((length(email) <= 255)),
    CONSTRAINT users_email_check1 CHECK ((email ~* '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'::text)),
    CONSTRAINT users_login_check CHECK ((length(login) <= 255)),
    CONSTRAINT users_password_check CHECK (((length(password) >= 8) AND (length(password) <= 255))),
    CONSTRAINT users_phone_check CHECK ((length(phone) <= 20)),
    CONSTRAINT users_phone_check1 CHECK ((phone ~ '^\+\d{10,15}$'::text))
);


--
-- Name: users_user_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.users ALTER COLUMN user_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.users_user_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: admins admins_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admins
    ADD CONSTRAINT admins_pkey PRIMARY KEY (admin_id);


--
-- Name: admins admins_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admins
    ADD CONSTRAINT admins_user_id_key UNIQUE (user_id);


--
-- Name: blacklist blacklist_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blacklist
    ADD CONSTRAINT blacklist_pkey PRIMARY KEY (block_id);


--
-- Name: chats chats_first_profile_id_second_profile_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chats
    ADD CONSTRAINT chats_first_profile_id_second_profile_id_key UNIQUE (first_profile_id, second_profile_id);


--
-- Name: chats chats_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chats
    ADD CONSTRAINT chats_pkey PRIMARY KEY (chat_id);


--
-- Name: complaint_types complaint_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.complaint_types
    ADD CONSTRAINT complaint_types_pkey PRIMARY KEY (comp_type);


--
-- Name: complaints complaints_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.complaints
    ADD CONSTRAINT complaints_pkey PRIMARY KEY (complaint_id);


--
-- Name: interests interests_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.interests
    ADD CONSTRAINT interests_pkey PRIMARY KEY (interest_id);


--
-- Name: likes likes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.likes
    ADD CONSTRAINT likes_pkey PRIMARY KEY (like_id);


--
-- Name: likes likes_profile_id_liked_profile_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.likes
    ADD CONSTRAINT likes_profile_id_liked_profile_id_key UNIQUE (profile_id, liked_profile_id);


--
-- Name: locations locations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locations
    ADD CONSTRAINT locations_pkey PRIMARY KEY (location_id);


--
-- Name: matches matches_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.matches
    ADD CONSTRAINT matches_pkey PRIMARY KEY (match_id);


--
-- Name: messages messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (message_id);


--
-- Name: notification_types notification_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_types
    ADD CONSTRAINT notification_types_pkey PRIMARY KEY (notif_type);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (notification_id);


--
-- Name: preferences preferences_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.preferences
    ADD CONSTRAINT preferences_pkey PRIMARY KEY (preference_id);


--
-- Name: profile_interests profile_interests_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_interests
    ADD CONSTRAINT profile_interests_pkey PRIMARY KEY (profile_id, interest_id);


--
-- Name: profile_preferences profile_preferences_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_preferences
    ADD CONSTRAINT profile_preferences_pkey PRIMARY KEY (profile_id, preference_id);


--
-- Name: profile_ratings profile_ratings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_ratings
    ADD CONSTRAINT profile_ratings_pkey PRIMARY KEY (rating_id);


--
-- Name: profiles profiles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profiles
    ADD CONSTRAINT profiles_pkey PRIMARY KEY (profile_id);


--
-- Name: queries queries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.queries
    ADD CONSTRAINT queries_pkey PRIMARY KEY (query_id);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);


--
-- Name: static static_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.static
    ADD CONSTRAINT static_pkey PRIMARY KEY (id);


--
-- Name: subscription_types subscription_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscription_types
    ADD CONSTRAINT subscription_types_pkey PRIMARY KEY (sub_type);


--
-- Name: subscriptions subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_pkey PRIMARY KEY (sub_id);


--
-- Name: blacklist unique_blacklist; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blacklist
    ADD CONSTRAINT unique_blacklist UNIQUE (user_id);


--
-- Name: matches unique_match; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.matches
    ADD CONSTRAINT unique_match UNIQUE (profile_id, matched_profile_id);


--
-- Name: user_answer user_answer_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_answer
    ADD CONSTRAINT user_answer_pkey PRIMARY KEY (answer_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_login_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_login_key UNIQUE (login);


--
-- Name: users users_phone_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_phone_key UNIQUE (phone);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (user_id);


--
-- Name: admins admins_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admins
    ADD CONSTRAINT admins_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: blacklist blacklist_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blacklist
    ADD CONSTRAINT blacklist_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: chats chats_first_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chats
    ADD CONSTRAINT chats_first_profile_id_fkey FOREIGN KEY (first_profile_id) REFERENCES public.profiles(profile_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: chats chats_last_sender_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chats
    ADD CONSTRAINT chats_last_sender_fkey FOREIGN KEY (last_sender) REFERENCES public.profiles(profile_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: chats chats_second_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chats
    ADD CONSTRAINT chats_second_profile_id_fkey FOREIGN KEY (second_profile_id) REFERENCES public.profiles(profile_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: complaints complaints_complaint_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.complaints
    ADD CONSTRAINT complaints_complaint_by_fkey FOREIGN KEY (complaint_by) REFERENCES public.users(user_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: complaints complaints_complaint_on_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.complaints
    ADD CONSTRAINT complaints_complaint_on_fkey FOREIGN KEY (complaint_on) REFERENCES public.users(user_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: complaints complaints_complaint_type_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.complaints
    ADD CONSTRAINT complaints_complaint_type_fkey FOREIGN KEY (complaint_type) REFERENCES public.complaint_types(comp_type) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: likes likes_liked_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.likes
    ADD CONSTRAINT likes_liked_profile_id_fkey FOREIGN KEY (liked_profile_id) REFERENCES public.profiles(profile_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: likes likes_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.likes
    ADD CONSTRAINT likes_profile_id_fkey FOREIGN KEY (profile_id) REFERENCES public.profiles(profile_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: matches matches_matched_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.matches
    ADD CONSTRAINT matches_matched_profile_id_fkey FOREIGN KEY (matched_profile_id) REFERENCES public.profiles(profile_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: matches matches_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.matches
    ADD CONSTRAINT matches_profile_id_fkey FOREIGN KEY (profile_id) REFERENCES public.profiles(profile_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: messages messages_chat_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_chat_id_fkey FOREIGN KEY (chat_id) REFERENCES public.chats(chat_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: messages messages_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.profiles(profile_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: notifications notifications_notification_type_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_notification_type_fkey FOREIGN KEY (notification_type) REFERENCES public.notification_types(notif_type) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: notifications notifications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: profile_interests profile_interests_interest_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_interests
    ADD CONSTRAINT profile_interests_interest_id_fkey FOREIGN KEY (interest_id) REFERENCES public.interests(interest_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: profile_interests profile_interests_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_interests
    ADD CONSTRAINT profile_interests_profile_id_fkey FOREIGN KEY (profile_id) REFERENCES public.profiles(profile_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: profile_preferences profile_preferences_preference_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_preferences
    ADD CONSTRAINT profile_preferences_preference_id_fkey FOREIGN KEY (preference_id) REFERENCES public.preferences(preference_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: profile_preferences profile_preferences_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_preferences
    ADD CONSTRAINT profile_preferences_profile_id_fkey FOREIGN KEY (profile_id) REFERENCES public.profiles(profile_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: profile_ratings profile_ratings_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_ratings
    ADD CONSTRAINT profile_ratings_profile_id_fkey FOREIGN KEY (profile_id) REFERENCES public.profiles(profile_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: profile_ratings profile_ratings_rated_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile_ratings
    ADD CONSTRAINT profile_ratings_rated_profile_id_fkey FOREIGN KEY (rated_profile_id) REFERENCES public.profiles(profile_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: profiles profiles_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profiles
    ADD CONSTRAINT profiles_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.locations(location_id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: sessions sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: static static_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.static
    ADD CONSTRAINT static_profile_id_fkey FOREIGN KEY (profile_id) REFERENCES public.profiles(profile_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: subscriptions subscriptions_sub_type_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_sub_type_fkey FOREIGN KEY (sub_type) REFERENCES public.subscription_types(sub_type) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: subscriptions subscriptions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_answer user_answer_query_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_answer
    ADD CONSTRAINT user_answer_query_id_fkey FOREIGN KEY (query_id) REFERENCES public.queries(query_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_answer user_answer_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_answer
    ADD CONSTRAINT user_answer_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: users users_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_profile_id_fkey FOREIGN KEY (profile_id) REFERENCES public.profiles(profile_id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: -
--

GRANT USAGE ON SCHEMA public TO app_user;


--
-- PostgreSQL database dump complete
--

