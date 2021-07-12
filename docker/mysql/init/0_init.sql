CREATE DATABASE IF NOT EXISTS sample_mysql CHARACTER SET utf8mb4;
CREATE DATABASE IF NOT EXISTS sample_mysql_test CHARACTER SET utf8mb4;
CREATE USER IF NOT EXISTS 'mysql_user'@'%' identified by 'mysql_pass';
GRANT ALL PRIVILEGES ON sample_mysql.* TO mysql_user@'%' WITH GRANT OPTION;
GRANT ALL PRIVILEGES ON sample_mysql_test.* TO mysql_user@'%' WITH GRANT OPTION;
FLUSH PRIVILEGES;
