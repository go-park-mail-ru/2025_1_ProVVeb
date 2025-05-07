package handlers

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Метрики для службы сообщений (чат)
var (
	// Чаты созданы
	messageChatsCreated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "message_chat_rooms_created_total",
			Help: "Total count of chat rooms created.",
		},
		[]string{"room_type"},
	)

	// Запросы просмотра комнат
	messageChatsViews = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "message_chat_room_views_total",
			Help: "Total views of chat room lists.",
		},
		[]string{},
	)

	// Удалённые комнаты
	messageChatsDeleted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "message_chat_rooms_deleted_total",
			Help: "Total deleted chat rooms.",
		},
		[]string{},
	)

	// Попыток отправить сообщение
	messageSent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "message_sent_total_attempts",
			Help: "Total messages attempts to send by users.",
		},
		[]string{},
	)

	// Принято сообщений
	messageReceived = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "message_received_total",
			Help: "Total messages received by clients.",
		},
		[]string{},
	)

	messageNotificationsFetched = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "message_notifications_fetched_total",
			Help: "Total notifications fetched by users.",
		},
	)
)

// profiles microservice metrics
var (
	profileRetrieved = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "profile_retrieved_total",
			Help: "Total profile retrieval operations.",
		},
		[]string{"action"},
	)

	profileUpdated = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "profile_updated_total",
			Help: "Total profile updates.",
		},
		[]string{"action"},
	)

	photoRemoved = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "photo_removed_total",
			Help: "Total photos removed from profiles.",
		},
		[]string{"action"},
	)

	likeSet = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "like_set_total",
			Help: "Total likes added to profiles.",
		},
		[]string{"action"},
	)

	searchPerformed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "search_performed_total",
			Help: "Total searches performed on profiles.",
		},
		[]string{"action"},
	)

	matchesRetrieved = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "matches_retrieved_total",
			Help: "Total retrieved matches between profiles.",
		},
		[]string{"action"},
	)

	photoUploaded = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "photo_uploaded_total",
			Help: "Total uploaded photos to profiles.",
		},
		[]string{"action"},
	)

	profilesListRetrieved = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "profiles_list_retrieved_total",
			Help: "Total retrievals of profile lists.",
		},
		[]string{"action"},
	)
)

// auth microservice metrics
var (
	loginAttempts = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "authentication_login_attempts_total",
			Help: "Total login attempts.",
		},
		[]string{"outcome"},
	)

	// Проверки сессии
	sessionChecks = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "authentication_session_checks_total",
			Help: "Total session checks.",
		},
		[]string{"status"},
	)

	// Попытки выхода из системы
	logoutAttempts = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "authentication_logout_attempts_total",
			Help: "Total logout attempts.",
		},
		[]string{"outcome"},
	)
)
