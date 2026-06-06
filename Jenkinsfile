pipeline {
    agent any

    environment {
    GOROOT = '/usr/local/go'
    GOPATH = '/var/lib/jenkins/go'
    PATH   = "/usr/local/go/bin:${env.PATH}"
    SONAR_TOKEN    = credentials('squ_3fb7160be8187a314e852ef03a4851b045d38780')
    VPS_SSH_KEY    = credentials('vps-ssh-key')
    SONAR_HOST_URL = 'http://95.111.228.35:9000'
    VPS_HOST       = '95.111.228.35'
    VPS_USER       = 'root'
}
    stages {

        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Test') {
            steps {
                sh '''
                    cd services/scraper-service && go test -coverprofile=coverage.out ./...
                    cd ../normalization && go test -coverprofile=coverage.out ./...
                    cd ../auth && go test -coverprofile=coverage.out ./...
                    cd ../history && go test -coverprofile=coverage.out ./...
                    cd ../proxy-validator && go test -coverprofile=coverage.out ./...
                    cd ../bff && go test -coverprofile=coverage.out ./...
                '''
            }
        }

        stage('SonarQube Analysis') {
            steps {
                withSonarQubeEnv('SonarQube') {
                    sh 'sonar-scanner'
                }
            }
        }

        stage('Quality Gate') {
            steps {
                timeout(time: 5, unit: 'MINUTES') {
                    waitForQualityGate abortPipeline: true
                }
            }
        }

        stage('Deploy') {
            when {
                branch 'main'
            }
            steps {
                sshagent(['vps-ssh-key']) {
                    sh '''
                        ssh -o StrictHostKeyChecking=no $VPS_USER@$VPS_HOST "
                            cd /opt/mergemarket &&
                            git pull origin main &&
                            docker compose pull &&
                            docker compose up -d --build &&
                            docker image prune -f
                        "
                    '''
                }
            }
        }

        stage('Service Tests') {
            steps {
                sh '''
                    for service in scraper-service normalization; do
                        if [ -f "services/$service/go.mod" ]; then
                            echo "Testing $service..."
                            cd services/$service
                            go test -coverprofile=coverage.out ./...
                            cd ../..
                        else
                            echo "Skipping $service — no go.mod found"
                        fi
                    done
                '''
            }
        }
    }

    post {
        success {
            echo 'Pipeline passed. Deployment complete.'
        }
        failure {
            echo 'Pipeline failed. Deployment blocked.'
        }
    }
}