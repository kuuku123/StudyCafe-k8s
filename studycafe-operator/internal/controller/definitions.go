package controller

import (
	corev1 "k8s.io/api/core/v1"
)

type appDef struct {
	name          string
	image         string
	port          int32
	needsJasypt   bool
	needsJwt      bool
	isHttpProbe   bool
	probePath     string
	memoryRequest string
	cpuRequest    string
	memoryLimit   string
	cpuLimit      string
	command       []string
}

type infraDef struct {
	name           string
	image          string
	ports          []int32
	env            []corev1.EnvVar
	pvcName        string
	pvcSize        string
	mountPath      string
	readinessCmd   []string
	livenessCmd    []string
	tcpProbePort   int32
	readinessDelay int32
	livenessDelay  int32
	memoryRequest  string
	cpuRequest     string
	memoryLimit    string
	cpuLimit       string
	hostname       string
	isStateful     bool
	command        []string
	args           []string
}

func getInfras() []infraDef {
	return []infraDef{
		{
			name:       "mysql-test",
			image:      "mysql:8.0",
			ports:      []int32{3306},
			isStateful: true,
			pvcName:    "mysql-data-pvc",
			pvcSize:    "1Gi",
			mountPath:  "/var/lib/mysql",
			env: []corev1.EnvVar{
				{Name: "MYSQL_DATABASE", Value: "test"},
				{
					Name: "MYSQL_ROOT_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "mysql-secret"},
							Key:                  "MYSQL_ROOT_PASSWORD",
						},
					},
				},
				{Name: "TZ", Value: "Asia/Seoul"},
			},
			readinessCmd:   []string{"mysqladmin", "ping", "-h", "localhost", "-u", "root", "-p1234"},
			livenessCmd:    []string{"mysqladmin", "ping", "-h", "localhost", "-u", "root", "-p1234"},
			readinessDelay: 180,
			livenessDelay:  300,
			memoryRequest:  "512Mi",
			cpuRequest:     "250m",
			memoryLimit:    "1Gi",
			cpuLimit:       "500m",
			args:           []string{"--character-set-server=utf8mb4", "--collation-server=utf8mb4_unicode_ci"},
		},
		{
			name:           "redis-test",
			image:          "redis:alpine",
			ports:          []int32{6379},
			isStateful:     true,
			pvcName:        "redis-data-pvc",
			pvcSize:        "512Mi",
			mountPath:      "/data",
			readinessCmd:   []string{"redis-cli", "ping"},
			livenessCmd:    []string{"redis-cli", "ping"},
			readinessDelay: 5,
			livenessDelay:  180,
			memoryRequest:  "128Mi",
			cpuRequest:     "100m",
			memoryLimit:    "256Mi",
			cpuLimit:       "250m",
		},
		{
			name:       "controller-1",
			hostname:   "controller-1",
			image:      "apache/kafka:latest",
			ports:      []int32{9093},
			isStateful: false,
			env: []corev1.EnvVar{
				{Name: "KAFKA_NODE_ID", Value: "1"},
				{Name: "KAFKA_PROCESS_ROLES", Value: "controller"},
				{Name: "KAFKA_LISTENERS", Value: "CONTROLLER://:9093"},
				{Name: "KAFKA_CONTROLLER_LISTENER_NAMES", Value: "CONTROLLER"},
				{Name: "KAFKA_CONTROLLER_QUORUM_VOTERS", Value: "1@controller-1:9093"},
				{Name: "KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS", Value: "0"},
				{Name: "CLUSTER_ID", Value: "DBwuQHeySamG6YAg1kmzcA"},
			},
			memoryRequest: "256Mi",
			cpuRequest:    "100m",
			memoryLimit:   "512Mi",
			cpuLimit:      "500m",
		},
		{
			name:       "broker-1",
			image:      "apache/kafka:latest",
			ports:      []int32{9092, 9093},
			isStateful: true,
			env: []corev1.EnvVar{
				{Name: "KAFKA_NODE_ID", Value: "2"},
				{Name: "KAFKA_PROCESS_ROLES", Value: "broker"},
				{Name: "KAFKA_LISTENERS", Value: "PLAINTEXT://:9092,PLAINTEXT_INTERNAL://:9093"},
				{Name: "KAFKA_ADVERTISED_LISTENERS", Value: "PLAINTEXT://broker-1:9092,PLAINTEXT_INTERNAL://broker-1:9093"},
				{Name: "KAFKA_INTER_BROKER_LISTENER_NAME", Value: "PLAINTEXT_INTERNAL"},
				{Name: "KAFKA_CONTROLLER_LISTENER_NAMES", Value: "CONTROLLER"},
				{Name: "KAFKA_LISTENER_SECURITY_PROTOCOL_MAP", Value: "CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT,PLAINTEXT_INTERNAL:PLAINTEXT"},
				{Name: "KAFKA_CONTROLLER_QUORUM_VOTERS", Value: "1@controller-1:9093"},
				{Name: "KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS", Value: "0"},
				{Name: "KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR", Value: "1"},
				{Name: "CLUSTER_ID", Value: "DBwuQHeySamG6YAg1kmzcA"},
			},
			tcpProbePort:   9092,
			readinessDelay: 30,
			memoryRequest:  "512Mi",
			cpuRequest:     "250m",
			memoryLimit:    "1Gi",
			cpuLimit:       "500m",
		},
	}
}

func getApps() []appDef {
	return []appDef{
		{
			name:          "api-gateway",
			image:         "kuuku123/api-gateway",
			port:          8083,
			needsJasypt:   false,
			needsJwt:      true,
			isHttpProbe:   false,
			memoryRequest: "256Mi",
			cpuRequest:    "250m",
			memoryLimit:   "512Mi",
			cpuLimit:      "500m",
			command:       nil,
		},
		{
			name:          "auth-service",
			image:         "kuuku123/auth-service",
			port:          8084,
			needsJasypt:   true,
			needsJwt:      true,
			isHttpProbe:   true,
			probePath:     "/actuator/health",
			memoryRequest: "512Mi",
			cpuRequest:    "250m",
			memoryLimit:   "1Gi",
			cpuLimit:      "1000m",
			command:       []string{"java", "-jar", "-Djasypt.encryptor.password=$(JASYPT_ENCRYPTOR_PASSWORD)", "-Dspring.profiles.active=prod", "/app/auth-service.jar"},
		},
		{
			name:          "study-service",
			image:         "kuuku123/study-service",
			port:          8081,
			needsJasypt:   true,
			needsJwt:      false,
			isHttpProbe:   true,
			probePath:     "/actuator/health",
			memoryRequest: "512Mi",
			cpuRequest:    "250m",
			memoryLimit:   "1Gi",
			cpuLimit:      "1000m",
			command:       []string{"java", "-jar", "-Djasypt.encryptor.password=$(JASYPT_ENCRYPTOR_PASSWORD)", "-Dspring.profiles.active=prod", "/app/study-service.jar"},
		},
		{
			name:          "notification-service",
			image:         "kuuku123/studycafe-webflux-notification",
			port:          8082,
			needsJasypt:   false,
			needsJwt:      false,
			isHttpProbe:   true,
			probePath:     "/actuator/health",
			memoryRequest: "256Mi",
			cpuRequest:    "250m",
			memoryLimit:   "512Mi",
			cpuLimit:      "500m",
			command:       nil,
		},
		{
			name:          "react-frontend",
			image:         "kuuku123/react-apache-app",
			port:          80,
			needsJasypt:   false,
			needsJwt:      false,
			isHttpProbe:   true,
			probePath:     "/",
			memoryRequest: "64Mi",
			cpuRequest:    "50m",
			memoryLimit:   "128Mi",
			cpuLimit:      "200m",
			command:       nil,
		},
	}
}
