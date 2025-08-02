// Package mybatis_tests SQL映射定义
//
// 定义所有SQL语句，展示MyBatis动态SQL的完整功能
package mybat

// SQL映射常量 - 基础CRUD操作
const (
	// ========== 基础查询 ==========
	SelectByIdSQL = `
		SELECT id, name, email, age, status, avatar, phone, birthday, 
		       created_at, updated_at, deleted_at 
		FROM users 
		WHERE id = #{id} AND deleted_at IS NULL
	`
	
	SelectByEmailSQL = `
		SELECT id, name, email, age, status, avatar, phone, birthday, 
		       created_at, updated_at, deleted_at 
		FROM users 
		WHERE email = #{email} AND deleted_at IS NULL
	`
	
	SelectByIdsSQL = `
		SELECT id, name, email, age, status, avatar, phone, birthday, 
		       created_at, updated_at, deleted_at 
		FROM users 
		WHERE id IN 
		<foreach collection="ids" item="id" open="(" separator="," close=")">
			#{id}
		</foreach>
		AND deleted_at IS NULL
		ORDER BY id
	`
	
	// ========== 动态查询 ==========
	SelectListSQL = `
		SELECT id, name, email, age, status, avatar, phone, birthday, 
		       created_at, updated_at, deleted_at 
		FROM users 
		<where>
			<if test="name != null and name != ''">
				AND name LIKE CONCAT('%', #{name}, '%')
			</if>
			<if test="email != null and email != ''">
				AND email = #{email}
			</if>
			<if test="status != null and status != ''">
				AND status = #{status}
			</if>
			<if test="phone != null and phone != ''">
				AND phone = #{phone}
			</if>
			<if test="ageMin > 0">
				AND age >= #{ageMin}
			</if>
			<if test="ageMax > 0">
				AND age <= #{ageMax}
			</if>
			<if test="createdAfter != null">
				AND created_at >= #{createdAfter}
			</if>
			<if test="createdBefore != null">
				AND created_at <= #{createdBefore}
			</if>
			<if test="keyword != null and keyword != ''">
				AND (name LIKE CONCAT('%', #{keyword}, '%') 
				     OR email LIKE CONCAT('%', #{keyword}, '%')
				     OR phone LIKE CONCAT('%', #{keyword}, '%'))
			</if>
			<if test="!includeDeleted">
				AND deleted_at IS NULL
			</if>
			<if test="onlyActive">
				AND status = 'active'
			</if>
		</where>
		<choose>
			<when test="orderBy != null and orderBy != ''">
				ORDER BY #{orderBy}
				<if test="orderDesc">DESC</if>
				<if test="!orderDesc">ASC</if>
			</when>
			<otherwise>
				ORDER BY created_at DESC
			</otherwise>
		</choose>
		<if test="pageSize > 0">
			LIMIT #{pageSize}
			<if test="page > 0">
				OFFSET #{offset}
			</if>
		</if>
	`
	
	SelectCountSQL = `
		SELECT COUNT(*) 
		FROM users 
		<where>
			<if test="name != null and name != ''">
				AND name LIKE CONCAT('%', #{name}, '%')
			</if>
			<if test="email != null and email != ''">
				AND email = #{email}
			</if>
			<if test="status != null and status != ''">
				AND status = #{status}
			</if>
			<if test="phone != null and phone != ''">
				AND phone = #{phone}
			</if>
			<if test="ageMin > 0">
				AND age >= #{ageMin}
			</if>
			<if test="ageMax > 0">
				AND age <= #{ageMax}
			</if>
			<if test="createdAfter != null">
				AND created_at >= #{createdAfter}
			</if>
			<if test="createdBefore != null">
				AND created_at <= #{createdBefore}
			</if>
			<if test="keyword != null and keyword != ''">
				AND (name LIKE CONCAT('%', #{keyword}, '%') 
				     OR email LIKE CONCAT('%', #{keyword}, '%')
				     OR phone LIKE CONCAT('%', #{keyword}, '%'))
			</if>
			<if test="!includeDeleted">
				AND deleted_at IS NULL
			</if>
			<if test="onlyActive">
				AND status = 'active'
			</if>
		</where>
	`
	
	// ========== 插入操作 ==========
	InsertSQL = `
		INSERT INTO users 
		(name, email, age, status, avatar, phone, birthday, created_at, updated_at)
		VALUES 
		(#{name}, #{email}, #{age}, #{status}, #{avatar}, #{phone}, #{birthday}, NOW(), NOW())
	`
	
	BatchInsertSQL = `
		INSERT INTO users 
		(name, email, age, status, avatar, phone, birthday, created_at, updated_at)
		VALUES
		<foreach collection="users" item="user" separator=",">
			(#{user.name}, #{user.email}, #{user.age}, #{user.status}, 
			 #{user.avatar}, #{user.phone}, #{user.birthday}, NOW(), NOW())
		</foreach>
	`
	
	// ========== 更新操作 ==========
	UpdateSQL = `
		UPDATE users 
		SET name = #{name}, 
		    email = #{email}, 
		    age = #{age}, 
		    status = #{status}, 
		    avatar = #{avatar}, 
		    phone = #{phone}, 
		    birthday = #{birthday}, 
		    updated_at = NOW()
		WHERE id = #{id}
	`
	
	UpdateSelectiveSQL = `
		UPDATE users 
		<set>
			<if test="name != null and name != ''">
				name = #{name},
			</if>
			<if test="email != null and email != ''">
				email = #{email},
			</if>
			<if test="age > 0">
				age = #{age},
			</if>
			<if test="status != null and status != ''">
				status = #{status},
			</if>
			<if test="avatar != null">
				avatar = #{avatar},
			</if>
			<if test="phone != null">
				phone = #{phone},
			</if>
			<if test="birthday != null">
				birthday = #{birthday},
			</if>
			updated_at = NOW()
		</set>
		WHERE id = #{id}
	`
	
	BatchUpdateSQL = `
		UPDATE users 
		<set>
			<foreach collection="updates" index="key" item="value" separator=",">
				${key} = #{value}
			</foreach>
			, updated_at = NOW()
		</set>
		WHERE id IN 
		<foreach collection="userIds" item="id" open="(" separator="," close=")">
			#{id}
		</foreach>
	`
	
	BatchUpdateStatusSQL = `
		UPDATE users 
		SET status = #{status}, updated_at = NOW()
		WHERE id IN 
		<foreach collection="ids" item="id" open="(" separator="," close=")">
			#{id}
		</foreach>
	`
	
	// ========== 删除操作 ==========
	DeleteSQL = `
		UPDATE users 
		SET deleted_at = NOW() 
		WHERE id = #{id}
	`
	
	PhysicalDeleteSQL = `
		DELETE FROM users 
		WHERE id = #{id}
	`
	
	BatchDeleteSQL = `
		UPDATE users 
		SET deleted_at = NOW() 
		WHERE id IN 
		<foreach collection="ids" item="id" open="(" separator="," close=")">
			#{id}
		</foreach>
	`
	
	// ========== 聚合查询 ==========
	SelectStatsSQL = `
		SELECT 
			COUNT(*) as total_users,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active_users,
			COUNT(CASE WHEN created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY) THEN 1 END) as recent_users
		FROM users 
		WHERE deleted_at IS NULL
	`
	
	SelectByStatusSQL = `
		SELECT 
			status as field,
			status as value,
			COUNT(*) as count
		FROM users 
		WHERE deleted_at IS NULL
		GROUP BY status
		ORDER BY count DESC
	`
	
	SelectByAgeGroupSQL = `
		SELECT 
			CASE 
				WHEN age < 18 THEN 'under_18'
				WHEN age BETWEEN 18 AND 25 THEN '18_25'
				WHEN age BETWEEN 26 AND 35 THEN '26_35'
				WHEN age BETWEEN 36 AND 50 THEN '36_50'
				ELSE 'over_50'
			END as field,
			CASE 
				WHEN age < 18 THEN 'under_18'
				WHEN age BETWEEN 18 AND 25 THEN '18_25'
				WHEN age BETWEEN 26 AND 35 THEN '26_35'
				WHEN age BETWEEN 36 AND 50 THEN '36_50'
				ELSE 'over_50'
			END as value,
			COUNT(*) as count
		FROM users 
		WHERE deleted_at IS NULL
		GROUP BY 
			CASE 
				WHEN age < 18 THEN 'under_18'
				WHEN age BETWEEN 18 AND 25 THEN '18_25'
				WHEN age BETWEEN 26 AND 35 THEN '26_35'
				WHEN age BETWEEN 36 AND 50 THEN '36_50'
				ELSE 'over_50'
			END
		ORDER BY count DESC
	`
	
	SelectActiveUsersInPeriodSQL = `
		SELECT id, name, email, age, status, avatar, phone, birthday, 
		       created_at, updated_at, deleted_at 
		FROM users 
		WHERE status = 'active' 
		  AND deleted_at IS NULL
		  AND created_at BETWEEN #{startTime} AND #{endTime}
		ORDER BY created_at DESC
	`
	
	// ========== 复杂查询 ==========
	SelectWithProfileSQL = `
		SELECT 
			u.id, u.name, u.email, u.age, u.status, u.avatar, u.phone, u.birthday,
			u.created_at, u.updated_at, u.deleted_at,
			p.bio, p.website, p.location, p.company, p.occupation, p.education,
			p.skills, p.preferences
		FROM users u
		LEFT JOIN user_profiles p ON u.id = p.user_id
		WHERE u.id = #{id} AND u.deleted_at IS NULL
	`
	
	SelectWithRolesSQL = `
		SELECT 
			u.id, u.name, u.email, u.age, u.status, u.avatar, u.phone, u.birthday,
			u.created_at, u.updated_at, u.deleted_at,
			r.id as role_id, r.role_name, r.permissions
		FROM users u
		LEFT JOIN user_roles r ON u.id = r.user_id
		WHERE u.id = #{id} AND u.deleted_at IS NULL
	`
	
	SelectWithArticlesSQL = `
		SELECT 
			u.id, u.name, u.email, u.age, u.status, u.avatar, u.phone, u.birthday,
			u.created_at, u.updated_at, u.deleted_at,
			a.id as article_id, a.title, a.summary, a.status as article_status,
			a.view_count, a.like_count, a.comment_count, a.published_at
		FROM users u
		LEFT JOIN articles a ON u.id = a.author_id AND a.deleted_at IS NULL
		WHERE u.id = #{userId} AND u.deleted_at IS NULL
		ORDER BY a.created_at DESC
		<if test="limit > 0">
			LIMIT #{limit}
		</if>
	`
	
	SearchUsersSQL = `
		SELECT id, name, email, age, status, avatar, phone, birthday, 
		       created_at, updated_at, deleted_at,
		       MATCH(name, email) AGAINST(#{keyword}) as relevance
		FROM users 
		WHERE deleted_at IS NULL
		  AND (MATCH(name, email) AGAINST(#{keyword})
		       OR name LIKE CONCAT('%', #{keyword}, '%')
		       OR email LIKE CONCAT('%', #{keyword}, '%')
		       OR phone LIKE CONCAT('%', #{keyword}, '%'))
		ORDER BY relevance DESC, created_at DESC
		<if test="limit > 0">
			LIMIT #{limit}
		</if>
	`
	
	SelectSimilarUsersSQL = `
		SELECT u2.id, u2.name, u2.email, u2.age, u2.status, u2.avatar, u2.phone, u2.birthday,
		       u2.created_at, u2.updated_at, u2.deleted_at
		FROM users u1
		JOIN users u2 ON u1.id != u2.id 
		  AND ABS(u1.age - u2.age) <= 5 
		  AND u1.status = u2.status
		WHERE u1.id = #{userId} 
		  AND u1.deleted_at IS NULL 
		  AND u2.deleted_at IS NULL
		ORDER BY ABS(u1.age - u2.age), u2.created_at DESC
		<if test="limit > 0">
			LIMIT #{limit}
		</if>
	`
	
	// ========== 特殊查询 ==========
	SelectRandomUsersSQL = `
		SELECT id, name, email, age, status, avatar, phone, birthday, 
		       created_at, updated_at, deleted_at 
		FROM users 
		WHERE deleted_at IS NULL AND status = 'active'
		ORDER BY RAND()
		<if test="limit > 0">
			LIMIT #{limit}
		</if>
	`
	
	SelectTopActiveUsersSQL = `
		SELECT u.id, u.name, u.email, u.age, u.status, u.avatar, u.phone, u.birthday,
		       u.created_at, u.updated_at, u.deleted_at,
		       COUNT(a.id) as article_count
		FROM users u
		LEFT JOIN articles a ON u.id = a.author_id AND a.deleted_at IS NULL
		WHERE u.deleted_at IS NULL AND u.status = 'active'
		GROUP BY u.id, u.name, u.email, u.age, u.status, u.avatar, u.phone, u.birthday,
		         u.created_at, u.updated_at, u.deleted_at
		ORDER BY article_count DESC, u.created_at DESC
		<if test="limit > 0">
			LIMIT #{limit}
		</if>
	`
	
	SelectUsersWithoutProfileSQL = `
		SELECT u.id, u.name, u.email, u.age, u.status, u.avatar, u.phone, u.birthday,
		       u.created_at, u.updated_at, u.deleted_at
		FROM users u
		LEFT JOIN user_profiles p ON u.id = p.user_id
		WHERE u.deleted_at IS NULL AND p.user_id IS NULL
		ORDER BY u.created_at DESC
	`
	
	SelectRecentRegistrationsSQL = `
		SELECT id, name, email, age, status, avatar, phone, birthday, 
		       created_at, updated_at, deleted_at 
		FROM users 
		WHERE deleted_at IS NULL 
		  AND created_at >= DATE_SUB(NOW(), INTERVAL #{days} DAY)
		ORDER BY created_at DESC
		<if test="limit > 0">
			LIMIT #{limit}
		</if>
	`
	
	// ========== 存储过程和函数 ==========
	CallUserStatsProcedureSQL = `
		CALL GetUserStats(#{startDate}, #{endDate})
	`
	
	SelectUserByCustomFunctionSQL = `
		SELECT id, name, email, age, status, avatar, phone, birthday, 
		       created_at, updated_at, deleted_at 
		FROM users 
		WHERE deleted_at IS NULL 
		  AND id IN (SELECT user_id FROM GetUsersByCustomFunction(#{param}))
		ORDER BY created_at DESC
	`
)

// 表创建SQL - 用于测试环境初始化
const (
	CreateUsersTableSQL = `
		CREATE TABLE IF NOT EXISTS users (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			age INT NOT NULL,
			status VARCHAR(20) DEFAULT 'active',
			avatar VARCHAR(255),
			phone VARCHAR(20),
			birthday DATE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL,
			
			INDEX idx_email (email),
			INDEX idx_status (status),
			INDEX idx_created_at (created_at),
			INDEX idx_deleted_at (deleted_at),
			FULLTEXT idx_search (name, email)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`
	
	CreateUserProfilesTableSQL = `
		CREATE TABLE IF NOT EXISTS user_profiles (
			user_id BIGINT PRIMARY KEY,
			bio TEXT,
			website VARCHAR(255),
			location VARCHAR(100),
			company VARCHAR(100),
			occupation VARCHAR(100),
			education VARCHAR(100),
			skills TEXT,
			preferences TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`
	
	CreateUserRolesTableSQL = `
		CREATE TABLE IF NOT EXISTS user_roles (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			user_id BIGINT NOT NULL,
			role_name VARCHAR(50) NOT NULL,
			permissions TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			
			INDEX idx_user_id (user_id),
			INDEX idx_role_name (role_name),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`
	
	CreateArticlesTableSQL = `
		CREATE TABLE IF NOT EXISTS articles (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			title VARCHAR(200) NOT NULL,
			content LONGTEXT,
			summary VARCHAR(500),
			author_id BIGINT NOT NULL,
			category_id BIGINT,
			tags VARCHAR(255),
			status VARCHAR(20) DEFAULT 'draft',
			view_count BIGINT DEFAULT 0,
			like_count BIGINT DEFAULT 0,
			comment_count BIGINT DEFAULT 0,
			published_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL,
			
			INDEX idx_author_id (author_id),
			INDEX idx_category_id (category_id),
			INDEX idx_status (status),
			INDEX idx_published_at (published_at),
			INDEX idx_created_at (created_at),
			INDEX idx_deleted_at (deleted_at),
			FULLTEXT idx_content_search (title, content, summary),
			
			FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`
	
	CreateCategoriesTableSQL = `
		CREATE TABLE IF NOT EXISTS categories (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(100) UNIQUE NOT NULL,
			slug VARCHAR(100) UNIQUE NOT NULL,
			description VARCHAR(500),
			parent_id BIGINT,
			sort_order INT DEFAULT 0,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			
			INDEX idx_parent_id (parent_id),
			INDEX idx_sort_order (sort_order),
			INDEX idx_is_active (is_active),
			FOREIGN KEY (parent_id) REFERENCES categories(id) ON DELETE SET NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`
	
	CreateUserArticleViewsTableSQL = `
		CREATE TABLE IF NOT EXISTS user_article_views (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			user_id BIGINT NOT NULL,
			article_id BIGINT NOT NULL,
			view_time TIMESTAMP NOT NULL,
			duration INT DEFAULT 0,
			device VARCHAR(50),
			user_agent VARCHAR(255),
			
			INDEX idx_user_id (user_id),
			INDEX idx_article_id (article_id),
			INDEX idx_view_time (view_time),
			UNIQUE KEY uk_user_article_time (user_id, article_id, view_time),
			
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`
)

// 测试数据SQL
const (
	InsertTestUsersSQL = `
		INSERT INTO users (name, email, age, status, avatar, phone, birthday, created_at, updated_at) VALUES
		('张三', 'zhangsan@example.com', 25, 'active', 'avatar1.jpg', '13800000001', '1998-01-15', NOW(), NOW()),
		('李四', 'lisi@example.com', 30, 'active', 'avatar2.jpg', '13800000002', '1993-05-20', NOW(), NOW()),
		('王五', 'wangwu@example.com', 28, 'inactive', 'avatar3.jpg', '13800000003', '1995-08-10', NOW(), NOW()),
		('赵六', 'zhaoliu@example.com', 35, 'active', 'avatar4.jpg', '13800000004', '1988-12-25', NOW(), NOW()),
		('钱七', 'qianqi@example.com', 22, 'banned', 'avatar5.jpg', '13800000005', '2001-03-08', NOW(), NOW()),
		('孙八', 'sunba@example.com', 40, 'active', 'avatar6.jpg', '13800000006', '1983-07-14', NOW(), NOW()),
		('周九', 'zhoujiu@example.com', 26, 'active', 'avatar7.jpg', '13800000007', '1997-11-30', NOW(), NOW()),
		('吴十', 'wushi@example.com', 33, 'inactive', 'avatar8.jpg', '13800000008', '1990-04-18', NOW(), NOW()),
		('郑十一', 'zhengshiyi@example.com', 29, 'active', 'avatar9.jpg', '13800000009', '1994-09-05', NOW(), NOW()),
		('王十二', 'wangshier@example.com', 31, 'active', 'avatar10.jpg', '13800000010', '1992-06-22', NOW(), NOW())
	`
	
	InsertTestProfilesSQL = `
		INSERT INTO user_profiles (user_id, bio, website, location, company, occupation, education, skills, preferences) VALUES
		(1, '热爱编程的软件工程师', 'https://zhangsan.dev', '北京', '阿里巴巴', '高级工程师', '清华大学', '["Go", "Java", "Python"]', '{"theme": "dark", "language": "zh"}'),
		(2, '产品经理，关注用户体验', 'https://lisi.pm', '上海', '腾讯', '产品经理', '北京大学', '["产品设计", "数据分析"]', '{"theme": "light", "language": "zh"}'),
		(4, '全栈开发者', 'https://zhaoliu.dev', '深圳', '字节跳动', '技术专家', '浙江大学', '["React", "Node.js", "MySQL"]', '{"theme": "auto", "language": "en"}')
	`
	
	InsertTestRolesSQL = `
		INSERT INTO user_roles (user_id, role_name, permissions) VALUES
		(1, 'admin', '["read", "write", "delete", "admin"]'),
		(1, 'developer', '["read", "write", "deploy"]'),
		(2, 'editor', '["read", "write"]'),
		(4, 'moderator', '["read", "write", "moderate"]')
	`
)

// 存储过程SQL
const (
	CreateUserStatsProcedureSQL = `
		DELIMITER //
		CREATE PROCEDURE IF NOT EXISTS GetUserStats(
			IN start_date DATE,
			IN end_date DATE
		)
		BEGIN
			SELECT 
				COUNT(*) as total_users,
				COUNT(CASE WHEN status = 'active' THEN 1 END) as active_users,
				COUNT(CASE WHEN created_at BETWEEN start_date AND end_date THEN 1 END) as recent_users
			FROM users 
			WHERE deleted_at IS NULL;
		END //
		DELIMITER ;
	`
	
	CreateCustomFunctionSQL = `
		DELIMITER //
		CREATE FUNCTION IF NOT EXISTS GetUsersByCustomFunction(param VARCHAR(255))
		RETURNS TEXT
		READS SQL DATA
		DETERMINISTIC
		BEGIN
			DECLARE result TEXT DEFAULT '';
			DECLARE done INT DEFAULT FALSE;
			DECLARE user_id BIGINT;
			DECLARE cur CURSOR FOR 
				SELECT id FROM users 
				WHERE deleted_at IS NULL 
				  AND (name LIKE CONCAT('%', param, '%') OR email LIKE CONCAT('%', param, '%'))
				LIMIT 10;
			DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = TRUE;
			
			OPEN cur;
			read_loop: LOOP
				FETCH cur INTO user_id;
				IF done THEN
					LEAVE read_loop;
				END IF;
				SET result = CONCAT(result, user_id, ',');
			END LOOP;
			CLOSE cur;
			
			RETURN TRIM(TRAILING ',' FROM result);
		END //
		DELIMITER ;
	`
)