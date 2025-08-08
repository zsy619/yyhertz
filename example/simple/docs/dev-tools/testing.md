# 测试工具

YYHertz 框架提供了完整的测试工具集，支持单元测试、集成测试、性能测试和端到端测试，帮助开发者确保应用程序的质量和稳定性。

## 概述

测试是软件开发的重要环节。YYHertz 的测试工具包含：

- 单元测试框架
- 集成测试支持
- Mock 和 Stub 工具
- 测试数据管理
- HTTP 测试客户端
- 数据库测试工具
- 性能基准测试
- 代码覆盖率分析

## 基本测试框架

### 单元测试

```go
package controllers_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "github.com/zsy619/yyhertz/framework/mvc/testing"
)

// 测试套件
type UserControllerTestSuite struct {
    suite.Suite
    app        *testing.TestApp
    controller *UserController
    userRepo   *mock.UserRepository
}

func (suite *UserControllerTestSuite) SetupSuite() {
    // 创建测试应用
    suite.app = testing.NewTestApp()
    
    // 创建 Mock 依赖
    suite.userRepo = &mock.UserRepository{}
    
    // 创建控制器
    suite.controller = &UserController{
        userRepo: suite.userRepo,
    }
    
    // 注册路由
    suite.app.AutoRouters(suite.controller)
}

func (suite *UserControllerTestSuite) TearDownSuite() {
    suite.app.Close()
}

func (suite *UserControllerTestSuite) SetupTest() {
    // 每个测试前重置 Mock
    suite.userRepo.Reset()
}

func (suite *UserControllerTestSuite) TestGetUser() {
    // 准备测试数据
    expectedUser := &User{
        ID:    1,
        Name:  "John Doe",
        Email: "john@example.com",
    }
    
    // 设置 Mock 期望
    suite.userRepo.On("GetByID", 1).Return(expectedUser, nil)
    
    // 执行请求
    resp := suite.app.GET("/users/1")
    
    // 断言响应
    assert.Equal(suite.T(), 200, resp.StatusCode)
    
    var user User
    resp.JSON(&user)
    assert.Equal(suite.T(), expectedUser.ID, user.ID)
    assert.Equal(suite.T(), expectedUser.Name, user.Name)
    assert.Equal(suite.T(), expectedUser.Email, user.Email)
    
    // 验证 Mock 调用
    suite.userRepo.AssertCalled(suite.T(), "GetByID", 1)
}

func (suite *UserControllerTestSuite) TestCreateUser() {
    // 准备请求数据
    request := CreateUserRequest{
        Name:  "Jane Doe",
        Email: "jane@example.com",
    }
    
    expectedUser := &User{
        ID:    2,
        Name:  request.Name,
        Email: request.Email,
    }
    
    // 设置 Mock 期望
    suite.userRepo.On("Create", mock.MatchedBy(func(user *User) bool {
        return user.Name == request.Name && user.Email == request.Email
    })).Return(expectedUser, nil)
    
    // 执行请求
    resp := suite.app.POST("/users").JSON(request)
    
    // 断言响应
    assert.Equal(suite.T(), 201, resp.StatusCode)
    
    var user User
    resp.JSON(&user)
    assert.Equal(suite.T(), expectedUser.ID, user.ID)
    assert.Equal(suite.T(), expectedUser.Name, user.Name)
    
    // 验证 Mock 调用
    suite.userRepo.AssertExpectations(suite.T())
}

func (suite *UserControllerTestSuite) TestGetUserNotFound() {
    // 设置 Mock 返回错误
    suite.userRepo.On("GetByID", 999).Return(nil, errors.New("user not found"))
    
    // 执行请求
    resp := suite.app.GET("/users/999")
    
    // 断言响应
    assert.Equal(suite.T(), 404, resp.StatusCode)
    
    var errorResp map[string]string
    resp.JSON(&errorResp)
    assert.Contains(suite.T(), errorResp["error"], "not found")
}

func TestUserControllerSuite(t *testing.T) {
    suite.Run(t, new(UserControllerTestSuite))
}
```

### 测试应用框架

```go
package testing

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "github.com/zsy619/yyhertz/framework/mvc"
)

type TestApp struct {
    app    *mvc.App
    server *httptest.Server
}

func NewTestApp() *TestApp {
    app := mvc.NewApp()
    
    // 配置测试环境
    app.SetMode("test")
    
    // 禁用日志输出
    app.SetLogLevel("error")
    
    return &TestApp{
        app: app,
    }
}

func (ta *TestApp) AutoRouters(controllers ...interface{}) {
    ta.app.AutoRouters(controllers...)
}

func (ta *TestApp) Start() {
    ta.server = httptest.NewServer(ta.app.Handler())
}

func (ta *TestApp) Close() {
    if ta.server != nil {
        ta.server.Close()
    }
}

func (ta *TestApp) GET(path string) *TestResponse {
    return ta.request("GET", path, nil)
}

func (ta *TestApp) POST(path string) *TestRequest {
    return &TestRequest{
        app:    ta,
        method: "POST",
        path:   path,
    }
}

func (ta *TestApp) PUT(path string) *TestRequest {
    return &TestRequest{
        app:    ta,
        method: "PUT",
        path:   path,
    }
}

func (ta *TestApp) DELETE(path string) *TestResponse {
    return ta.request("DELETE", path, nil)
}

func (ta *TestApp) request(method, path string, body []byte) *TestResponse {
    if ta.server == nil {
        ta.Start()
    }
    
    url := ta.server.URL + path
    
    var req *http.Request
    var err error
    
    if body != nil {
        req, err = http.NewRequest(method, url, bytes.NewBuffer(body))
    } else {
        req, err = http.NewRequest(method, url, nil)
    }
    
    if err != nil {
        panic(err)
    }
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    
    return &TestResponse{
        Response: resp,
    }
}

type TestRequest struct {
    app     *TestApp
    method  string
    path    string
    headers map[string]string
    query   map[string]string
}

func (tr *TestRequest) Header(key, value string) *TestRequest {
    if tr.headers == nil {
        tr.headers = make(map[string]string)
    }
    tr.headers[key] = value
    return tr
}

func (tr *TestRequest) Query(key, value string) *TestRequest {
    if tr.query == nil {
        tr.query = make(map[string]string)
    }
    tr.query[key] = value
    return tr
}

func (tr *TestRequest) JSON(data interface{}) *TestResponse {
    jsonData, err := json.Marshal(data)
    if err != nil {
        panic(err)
    }
    
    return tr.app.request(tr.method, tr.path, jsonData)
}

func (tr *TestRequest) Form(data map[string]string) *TestResponse {
    // 实现表单数据提交
    // ...
    return nil
}

type TestResponse struct {
    *http.Response
}

func (tr *TestResponse) JSON(dest interface{}) error {
    defer tr.Body.Close()
    return json.NewDecoder(tr.Body).Decode(dest)
}

func (tr *TestResponse) String() (string, error) {
    defer tr.Body.Close()
    buf := new(bytes.Buffer)
    buf.ReadFrom(tr.Body)
    return buf.String(), nil
}
```

## Mock 工具

### Mock 生成器

```go
//go:generate mockery --name=UserRepository --output=mocks --outpkg=mocks

package mock

import (
    "github.com/stretchr/testify/mock"
)

// Mock 用户仓库
type UserRepository struct {
    mock.Mock
}

func (m *UserRepository) GetByID(id int) (*User, error) {
    args := m.Called(id)
    return args.Get(0).(*User), args.Error(1)
}

func (m *UserRepository) Create(user *User) (*User, error) {
    args := m.Called(user)
    return args.Get(0).(*User), args.Error(1)
}

func (m *UserRepository) Update(user *User) error {
    args := m.Called(user)
    return args.Error(0)
}

func (m *UserRepository) Delete(id int) error {
    args := m.Called(id)
    return args.Error(0)
}

func (m *UserRepository) List(filters UserFilters) ([]*User, error) {
    args := m.Called(filters)
    return args.Get(0).([]*User), args.Error(1)
}

func (m *UserRepository) Reset() {
    m.ExpectedCalls = nil
    m.Calls = nil
}

// Mock HTTP 客户端
type HTTPClient struct {
    mock.Mock
}

func (m *HTTPClient) Get(url string) (*http.Response, error) {
    args := m.Called(url)
    return args.Get(0).(*http.Response), args.Error(1)
}

func (m *HTTPClient) Post(url string, data interface{}) (*http.Response, error) {
    args := m.Called(url, data)
    return args.Get(0).(*http.Response), args.Error(1)
}

// Mock 缓存
type Cache struct {
    mock.Mock
}

func (m *Cache) Get(key string) (interface{}, bool) {
    args := m.Called(key)
    return args.Get(0), args.Bool(1)
}

func (m *Cache) Set(key string, value interface{}, ttl time.Duration) error {
    args := m.Called(key, value, ttl)
    return args.Error(0)
}

func (m *Cache) Delete(key string) error {
    args := m.Called(key)
    return args.Error(0)
}
```

### 高级 Mock 功能

```go
package testing

// Mock 构建器
type MockBuilder struct {
    mocks map[string]interface{}
}

func NewMockBuilder() *MockBuilder {
    return &MockBuilder{
        mocks: make(map[string]interface{}),
    }
}

func (mb *MockBuilder) WithUserRepo() *MockBuilder {
    mb.mocks["userRepo"] = &mock.UserRepository{}
    return mb
}

func (mb *MockBuilder) WithCache() *MockBuilder {
    mb.mocks["cache"] = &mock.Cache{}
    return mb
}

func (mb *MockBuilder) WithHTTPClient() *MockBuilder {
    mb.mocks["httpClient"] = &mock.HTTPClient{}
    return mb
}

func (mb *MockBuilder) Build() map[string]interface{} {
    return mb.mocks
}

// 预设 Mock 行为
type MockPresets struct {
    userRepo *mock.UserRepository
}

func NewMockPresets(mocks map[string]interface{}) *MockPresets {
    return &MockPresets{
        userRepo: mocks["userRepo"].(*mock.UserRepository),
    }
}

func (mp *MockPresets) UserExists(id int, user *User) {
    mp.userRepo.On("GetByID", id).Return(user, nil)
}

func (mp *MockPresets) UserNotFound(id int) {
    mp.userRepo.On("GetByID", id).Return(nil, errors.New("user not found"))
}

func (mp *MockPresets) CreateUserSuccess(user *User) {
    mp.userRepo.On("Create", mock.AnythingOfType("*User")).Return(user, nil)
}

func (mp *MockPresets) CreateUserFails(err error) {
    mp.userRepo.On("Create", mock.AnythingOfType("*User")).Return(nil, err)
}

// 使用示例
func TestUserService(t *testing.T) {
    // 构建 Mock 对象
    mocks := NewMockBuilder().
        WithUserRepo().
        WithCache().
        Build()
    
    // 设置预设行为
    presets := NewMockPresets(mocks)
    presets.UserExists(1, &User{ID: 1, Name: "John"})
    presets.CreateUserSuccess(&User{ID: 2, Name: "Jane"})
    
    // 创建服务
    service := NewUserService(
        mocks["userRepo"].(UserRepository),
        mocks["cache"].(Cache),
    )
    
    // 执行测试
    user, err := service.GetUser(1)
    assert.NoError(t, err)
    assert.Equal(t, "John", user.Name)
}
```

## 数据库测试

### 测试数据库设置

```go
package testing

import (
    "database/sql"
    "github.com/DATA-DOG/go-txdb"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

func init() {
    // 注册事务数据库驱动
    txdb.Register("txdb", "mysql", "user:pass@tcp(localhost:3306)/testdb?parseTime=true")
}

type DBTestSuite struct {
    suite.Suite
    db   *gorm.DB
    conn *sql.DB
}

func (suite *DBTestSuite) SetupTest() {
    // 每个测试使用独立的事务
    conn, err := sql.Open("txdb", "test")
    suite.Require().NoError(err)
    
    db, err := gorm.Open(mysql.New(mysql.Config{
        Conn: conn,
    }), &gorm.Config{})
    suite.Require().NoError(err)
    
    suite.db = db
    suite.conn = conn
    
    // 迁移数据库
    suite.migrateDatabase()
    
    // 插入测试数据
    suite.seedDatabase()
}

func (suite *DBTestSuite) TearDownTest() {
    suite.conn.Close()
}

func (suite *DBTestSuite) migrateDatabase() {
    suite.db.AutoMigrate(&User{}, &Post{}, &Comment{})
}

func (suite *DBTestSuite) seedDatabase() {
    users := []User{
        {Name: "John Doe", Email: "john@example.com"},
        {Name: "Jane Smith", Email: "jane@example.com"},
    }
    
    for _, user := range users {
        suite.db.Create(&user)
    }
}

func (suite *DBTestSuite) TestUserRepository() {
    repo := NewUserRepository(suite.db)
    
    // 测试创建用户
    user := &User{
        Name:  "Test User",
        Email: "test@example.com",
    }
    
    err := repo.Create(user)
    suite.NoError(err)
    suite.NotZero(user.ID)
    
    // 测试查询用户
    found, err := repo.GetByID(user.ID)
    suite.NoError(err)
    suite.Equal(user.Name, found.Name)
    suite.Equal(user.Email, found.Email)
    
    // 测试更新用户
    found.Name = "Updated Name"
    err = repo.Update(found)
    suite.NoError(err)
    
    updated, err := repo.GetByID(user.ID)
    suite.NoError(err)
    suite.Equal("Updated Name", updated.Name)
    
    // 测试删除用户
    err = repo.Delete(user.ID)
    suite.NoError(err)
    
    _, err = repo.GetByID(user.ID)
    suite.Error(err)
}

func TestDBSuite(t *testing.T) {
    suite.Run(t, new(DBTestSuite))
}
```

### 测试数据工厂

```go
package factory

import (
    "time"
    "github.com/brianvoe/gofakeit/v6"
)

// 用户工厂
type UserFactory struct {
    id       int
    name     string
    email    string
    age      int
    active   bool
    createAt time.Time
}

func NewUser() *UserFactory {
    return &UserFactory{
        name:     gofakeit.Name(),
        email:    gofakeit.Email(),
        age:      gofakeit.Number(18, 65),
        active:   true,
        createAt: time.Now(),
    }
}

func (f *UserFactory) WithID(id int) *UserFactory {
    f.id = id
    return f
}

func (f *UserFactory) WithName(name string) *UserFactory {
    f.name = name
    return f
}

func (f *UserFactory) WithEmail(email string) *UserFactory {
    f.email = email
    return f
}

func (f *UserFactory) WithAge(age int) *UserFactory {
    f.age = age
    return f
}

func (f *UserFactory) Inactive() *UserFactory {
    f.active = false
    return f
}

func (f *UserFactory) Build() *User {
    return &User{
        ID:       f.id,
        Name:     f.name,
        Email:    f.email,
        Age:      f.age,
        Active:   f.active,
        CreateAt: f.createAt,
    }
}

func (f *UserFactory) Create(db *gorm.DB) *User {
    user := f.Build()
    db.Create(user)
    return user
}

// 批量创建
func (f *UserFactory) CreateBatch(db *gorm.DB, count int) []*User {
    var users []*User
    for i := 0; i < count; i++ {
        user := NewUser().Build()
        users = append(users, user)
    }
    db.Create(&users)
    return users
}

// 使用示例
func TestUserService(t *testing.T) {
    // 创建测试数据
    user := factory.NewUser().
        WithName("John Doe").
        WithEmail("john@example.com").
        WithAge(25).
        Create(db)
    
    // 批量创建
    users := factory.NewUser().CreateBatch(db, 10)
    
    // 创建特定状态的用户
    inactiveUser := factory.NewUser().
        Inactive().
        Create(db)
}

// 复杂对象工厂
type PostFactory struct {
    title    string
    content  string
    authorID int
    tags     []string
    status   string
}

func NewPost() *PostFactory {
    return &PostFactory{
        title:   gofakeit.Sentence(5),
        content: gofakeit.Paragraph(3, 5, 10, " "),
        tags:    []string{gofakeit.Word(), gofakeit.Word()},
        status:  "published",
    }
}

func (f *PostFactory) WithAuthor(author *User) *PostFactory {
    f.authorID = author.ID
    return f
}

func (f *PostFactory) WithTitle(title string) *PostFactory {
    f.title = title
    return f
}

func (f *PostFactory) Draft() *PostFactory {
    f.status = "draft"
    return f
}

func (f *PostFactory) Build() *Post {
    return &Post{
        Title:    f.title,
        Content:  f.content,
        AuthorID: f.authorID,
        Tags:     f.tags,
        Status:   f.status,
    }
}
```

## 集成测试

### HTTP 集成测试

```go
package integration_test

import (
    "testing"
    "github.com/stretchr/testify/suite"
    "github.com/zsy619/yyhertz/framework/mvc/testing"
)

type IntegrationTestSuite struct {
    suite.Suite
    app    *testing.TestApp
    db     *gorm.DB
    cache  cache.Cache
}

func (suite *IntegrationTestSuite) SetupSuite() {
    // 创建测试应用
    suite.app = testing.NewTestApp()
    
    // 连接测试数据库
    suite.db = setupTestDatabase()
    
    // 连接测试缓存
    suite.cache = setupTestCache()
    
    // 配置依赖注入
    suite.app.Container.Bind("db", suite.db)
    suite.app.Container.Bind("cache", suite.cache)
    
    // 注册控制器
    suite.app.AutoRouters(
        &UserController{},
        &PostController{},
        &AuthController{},
    )
    
    // 启动测试服务器
    suite.app.Start()
}

func (suite *IntegrationTestSuite) TearDownSuite() {
    suite.app.Close()
    cleanupTestDatabase(suite.db)
}

func (suite *IntegrationTestSuite) SetupTest() {
    // 清理数据库
    suite.cleanDatabase()
    
    // 清理缓存
    suite.cache.Clear()
    
    // 插入基础测试数据
    suite.seedDatabase()
}

func (suite *IntegrationTestSuite) TestUserCRUD() {
    // 测试创建用户
    userData := map[string]interface{}{
        "name":  "John Doe",
        "email": "john@example.com",
        "age":   25,
    }
    
    createResp := suite.app.POST("/api/users").JSON(userData)
    suite.Equal(201, createResp.StatusCode)
    
    var createdUser User
    createResp.JSON(&createdUser)
    suite.NotZero(createdUser.ID)
    suite.Equal(userData["name"], createdUser.Name)
    
    // 测试获取用户
    getResp := suite.app.GET(fmt.Sprintf("/api/users/%d", createdUser.ID))
    suite.Equal(200, getResp.StatusCode)
    
    var fetchedUser User
    getResp.JSON(&fetchedUser)
    suite.Equal(createdUser.ID, fetchedUser.ID)
    suite.Equal(createdUser.Name, fetchedUser.Name)
    
    // 测试更新用户
    updateData := map[string]interface{}{
        "name": "John Smith",
        "age":  26,
    }
    
    updateResp := suite.app.PUT(fmt.Sprintf("/api/users/%d", createdUser.ID)).JSON(updateData)
    suite.Equal(200, updateResp.StatusCode)
    
    var updatedUser User
    updateResp.JSON(&updatedUser)
    suite.Equal("John Smith", updatedUser.Name)
    suite.Equal(26, updatedUser.Age)
    
    // 测试删除用户
    deleteResp := suite.app.DELETE(fmt.Sprintf("/api/users/%d", createdUser.ID))
    suite.Equal(204, deleteResp.StatusCode)
    
    // 确认用户已删除
    getResp2 := suite.app.GET(fmt.Sprintf("/api/users/%d", createdUser.ID))
    suite.Equal(404, getResp2.StatusCode)
}

func (suite *IntegrationTestSuite) TestUserAuthentication() {
    // 创建用户
    user := factory.NewUser().
        WithEmail("test@example.com").
        WithPassword("password123").
        Create(suite.db)
    
    // 测试登录
    loginData := map[string]string{
        "email":    "test@example.com",
        "password": "password123",
    }
    
    loginResp := suite.app.POST("/api/auth/login").JSON(loginData)
    suite.Equal(200, loginResp.StatusCode)
    
    var loginResult map[string]interface{}
    loginResp.JSON(&loginResult)
    suite.Contains(loginResult, "token")
    
    token := loginResult["token"].(string)
    
    // 测试使用 token 访问受保护的资源
    protectedResp := suite.app.GET("/api/user/profile").
        Header("Authorization", "Bearer "+token)
    suite.Equal(200, protectedResp.StatusCode)
    
    // 测试无效 token
    invalidResp := suite.app.GET("/api/user/profile").
        Header("Authorization", "Bearer invalid_token")
    suite.Equal(401, invalidResp.StatusCode)
    
    // 测试登出
    logoutResp := suite.app.POST("/api/auth/logout").
        Header("Authorization", "Bearer "+token)
    suite.Equal(200, logoutResp.StatusCode)
    
    // 确认 token 失效
    afterLogoutResp := suite.app.GET("/api/user/profile").
        Header("Authorization", "Bearer "+token)
    suite.Equal(401, afterLogoutResp.StatusCode)
}

func (suite *IntegrationTestSuite) TestCacheIntegration() {
    // 创建用户
    user := factory.NewUser().Create(suite.db)
    
    // 第一次请求（应该从数据库获取）
    resp1 := suite.app.GET(fmt.Sprintf("/api/users/%d", user.ID))
    suite.Equal(200, resp1.StatusCode)
    
    // 验证缓存中有数据
    cacheKey := fmt.Sprintf("user:%d", user.ID)
    cachedUser, exists := suite.cache.Get(cacheKey)
    suite.True(exists)
    suite.Equal(user.ID, cachedUser.(*User).ID)
    
    // 第二次请求（应该从缓存获取）
    resp2 := suite.app.GET(fmt.Sprintf("/api/users/%d", user.ID))
    suite.Equal(200, resp2.StatusCode)
    
    // 更新用户（应该清除缓存）
    updateData := map[string]interface{}{"name": "Updated Name"}
    updateResp := suite.app.PUT(fmt.Sprintf("/api/users/%d", user.ID)).JSON(updateData)
    suite.Equal(200, updateResp.StatusCode)
    
    // 验证缓存已清除
    _, exists = suite.cache.Get(cacheKey)
    suite.False(exists)
}

func TestIntegrationSuite(t *testing.T) {
    suite.Run(t, new(IntegrationTestSuite))
}
```

## 性能测试

### 基准测试

```go
package benchmark_test

import (
    "testing"
    "github.com/zsy619/yyhertz/framework/mvc/testing"
)

func BenchmarkUserAPI(b *testing.B) {
    app := testing.NewTestApp()
    defer app.Close()
    
    // 设置测试数据
    setupBenchmarkData(app)
    
    app.Start()
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        resp := app.GET("/api/users/1")
        if resp.StatusCode != 200 {
            b.Fatalf("Expected status 200, got %d", resp.StatusCode)
        }
    }
}

func BenchmarkUserCreation(b *testing.B) {
    app := testing.NewTestApp()
    defer app.Close()
    
    app.Start()
    
    userData := map[string]interface{}{
        "name":  "Benchmark User",
        "email": "bench@example.com",
    }
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        userData["email"] = fmt.Sprintf("bench%d@example.com", i)
        resp := app.POST("/api/users").JSON(userData)
        if resp.StatusCode != 201 {
            b.Fatalf("Expected status 201, got %d", resp.StatusCode)
        }
    }
}

// 并发性能测试
func BenchmarkUserAPIConcurrent(b *testing.B) {
    app := testing.NewTestApp()
    defer app.Close()
    
    setupBenchmarkData(app)
    app.Start()
    
    b.ResetTimer()
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            resp := app.GET("/api/users/1")
            if resp.StatusCode != 200 {
                b.Fatalf("Expected status 200, got %d", resp.StatusCode)
            }
        }
    })
}

// 内存分配基准测试
func BenchmarkUserServiceMemory(b *testing.B) {
    service := NewUserService(mockRepo, mockCache)
    
    b.ReportAllocs()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        user, err := service.GetUser(1)
        if err != nil {
            b.Fatal(err)
        }
        _ = user
    }
}

// 数据库性能测试
func BenchmarkDatabaseOperations(b *testing.B) {
    db := setupTestDatabase()
    defer cleanupTestDatabase(db)
    
    repo := NewUserRepository(db)
    
    b.Run("Insert", func(b *testing.B) {
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            user := &User{
                Name:  fmt.Sprintf("User%d", i),
                Email: fmt.Sprintf("user%d@example.com", i),
            }
            err := repo.Create(user)
            if err != nil {
                b.Fatal(err)
            }
        }
    })
    
    b.Run("Select", func(b *testing.B) {
        // 先插入一些数据
        for i := 0; i < 1000; i++ {
            user := &User{
                Name:  fmt.Sprintf("User%d", i),
                Email: fmt.Sprintf("user%d@example.com", i),
            }
            repo.Create(user)
        }
        
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            _, err := repo.GetByID(i%1000 + 1)
            if err != nil {
                b.Fatal(err)
            }
        }
    })
}
```

### 负载测试

```go
package loadtest

import (
    "context"
    "sync"
    "time"
    "github.com/zsy619/yyhertz/framework/mvc/testing"
)

type LoadTestConfig struct {
    Concurrency int
    Duration    time.Duration
    RPS         int // 每秒请求数
}

type LoadTestResult struct {
    TotalRequests   int
    SuccessRequests int
    FailedRequests  int
    AverageLatency  time.Duration
    MaxLatency      time.Duration
    MinLatency      time.Duration
    Throughput      float64 // 请求/秒
    ErrorRate       float64
}

type LoadTester struct {
    app    *testing.TestApp
    config LoadTestConfig
    
    results    []RequestResult
    resultsMux sync.Mutex
}

type RequestResult struct {
    Latency    time.Duration
    StatusCode int
    Error      error
}

func NewLoadTester(app *testing.TestApp, config LoadTestConfig) *LoadTester {
    return &LoadTester{
        app:     app,
        config:  config,
        results: make([]RequestResult, 0),
    }
}

func (lt *LoadTester) Run(endpoint string) *LoadTestResult {
    ctx, cancel := context.WithTimeout(context.Background(), lt.config.Duration)
    defer cancel()
    
    var wg sync.WaitGroup
    
    // 计算请求间隔
    requestInterval := time.Second / time.Duration(lt.config.RPS)
    
    // 启动并发 workers
    for i := 0; i < lt.config.Concurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            lt.worker(ctx, endpoint, requestInterval)
        }()
    }
    
    wg.Wait()
    
    return lt.calculateResults()
}

func (lt *LoadTester) worker(ctx context.Context, endpoint string, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            lt.makeRequest(endpoint)
        }
    }
}

func (lt *LoadTester) makeRequest(endpoint string) {
    start := time.Now()
    
    resp := lt.app.GET(endpoint)
    latency := time.Since(start)
    
    result := RequestResult{
        Latency:    latency,
        StatusCode: resp.StatusCode,
    }
    
    lt.resultsMux.Lock()
    lt.results = append(lt.results, result)
    lt.resultsMux.Unlock()
}

func (lt *LoadTester) calculateResults() *LoadTestResult {
    lt.resultsMux.Lock()
    defer lt.resultsMux.Unlock()
    
    if len(lt.results) == 0 {
        return &LoadTestResult{}
    }
    
    var totalLatency time.Duration
    var successCount, failedCount int
    var maxLatency, minLatency time.Duration
    
    minLatency = lt.results[0].Latency
    
    for _, result := range lt.results {
        totalLatency += result.Latency
        
        if result.StatusCode >= 200 && result.StatusCode < 400 {
            successCount++
        } else {
            failedCount++
        }
        
        if result.Latency > maxLatency {
            maxLatency = result.Latency
        }
        
        if result.Latency < minLatency {
            minLatency = result.Latency
        }
    }
    
    totalRequests := len(lt.results)
    averageLatency := totalLatency / time.Duration(totalRequests)
    throughput := float64(totalRequests) / lt.config.Duration.Seconds()
    errorRate := float64(failedCount) / float64(totalRequests)
    
    return &LoadTestResult{
        TotalRequests:   totalRequests,
        SuccessRequests: successCount,
        FailedRequests:  failedCount,
        AverageLatency:  averageLatency,
        MaxLatency:      maxLatency,
        MinLatency:      minLatency,
        Throughput:      throughput,
        ErrorRate:       errorRate,
    }
}

// 使用示例
func TestLoadAPI(t *testing.T) {
    app := testing.NewTestApp()
    defer app.Close()
    
    // 设置测试数据
    setupLoadTestData(app)
    app.Start()
    
    config := LoadTestConfig{
        Concurrency: 10,
        Duration:    30 * time.Second,
        RPS:         100,
    }
    
    tester := NewLoadTester(app, config)
    result := tester.Run("/api/users/1")
    
    // 验证结果
    assert.True(t, result.ErrorRate < 0.01) // 错误率低于1%
    assert.True(t, result.AverageLatency < 100*time.Millisecond) // 平均延迟低于100ms
    assert.True(t, result.Throughput > 90) // 吞吐量超过90 RPS
    
    t.Logf("Load Test Results:")
    t.Logf("Total Requests: %d", result.TotalRequests)
    t.Logf("Success Rate: %.2f%%", float64(result.SuccessRequests)/float64(result.TotalRequests)*100)
    t.Logf("Average Latency: %v", result.AverageLatency)
    t.Logf("Max Latency: %v", result.MaxLatency)
    t.Logf("Throughput: %.2f RPS", result.Throughput)
    t.Logf("Error Rate: %.2f%%", result.ErrorRate*100)
}
```

## 代码覆盖率

### 覆盖率配置

```bash
#!/bin/bash
# coverage.sh

# 运行测试并生成覆盖率报告
go test -coverprofile=coverage.out ./...

# 生成 HTML 报告
go tool cover -html=coverage.out -o coverage.html

# 显示覆盖率摘要
go tool cover -func=coverage.out

# 检查覆盖率阈值
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
THRESHOLD=80

if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
    echo "Coverage $COVERAGE% is below threshold $THRESHOLD%"
    exit 1
else
    echo "Coverage $COVERAGE% meets threshold $THRESHOLD%"
fi
```

### Makefile 集成

```makefile
# Makefile

.PHONY: test test-unit test-integration test-load coverage benchmark

test: test-unit test-integration

test-unit:
	go test -v -race ./internal/...

test-integration:
	go test -v -tags=integration ./test/integration/...

test-load:
	go test -v -tags=load ./test/load/...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

benchmark:
	go test -bench=. -benchmem ./...

clean-test:
	rm -f coverage.out coverage.html
	docker-compose -f docker-compose.test.yml down -v

setup-test:
	docker-compose -f docker-compose.test.yml up -d
	sleep 5
	go run scripts/migrate.go
```

## 最佳实践

### 1. 测试组织

```go
// 按功能组织测试
package user_test

import (
    "testing"
    "github.com/stretchr/testify/suite"
)

// 单元测试套件
type UserUnitTestSuite struct {
    suite.Suite
    service *UserService
    mocks   map[string]interface{}
}

// 集成测试套件
type UserIntegrationTestSuite struct {
    suite.Suite
    app *testing.TestApp
    db  *gorm.DB
}

// 性能测试
func BenchmarkUserOperations(b *testing.B) {
    // 性能测试代码
}

// 使用构建标签分离测试
// +build unit
func TestUserService(t *testing.T) {
    suite.Run(t, new(UserUnitTestSuite))
}

// +build integration
func TestUserIntegration(t *testing.T) {
    suite.Run(t, new(UserIntegrationTestSuite))
}
```

### 2. 测试数据管理

```go
// 测试数据管理器
type TestDataManager struct {
    db      *gorm.DB
    cleanup []func()
}

func NewTestDataManager(db *gorm.DB) *TestDataManager {
    return &TestDataManager{
        db:      db,
        cleanup: make([]func(), 0),
    }
}

func (tdm *TestDataManager) CreateUser(attrs ...func(*User)) *User {
    user := factory.NewUser().Build()
    
    // 应用自定义属性
    for _, attr := range attrs {
        attr(user)
    }
    
    tdm.db.Create(user)
    
    // 注册清理函数
    tdm.cleanup = append(tdm.cleanup, func() {
        tdm.db.Delete(user)
    })
    
    return user
}

func (tdm *TestDataManager) Cleanup() {
    for i := len(tdm.cleanup) - 1; i >= 0; i-- {
        tdm.cleanup[i]()
    }
    tdm.cleanup = tdm.cleanup[:0]
}

// 使用示例
func TestUserOperations(t *testing.T) {
    dm := NewTestDataManager(db)
    defer dm.Cleanup()
    
    // 创建测试用户
    user := dm.CreateUser(func(u *User) {
        u.Name = "Test User"
        u.Age = 25
    })
    
    // 进行测试
    // ...
}
```

### 3. 测试工具函数

```go
// 测试助手函数
package testutil

// 断言 JSON 响应
func AssertJSONResponse(t *testing.T, resp *TestResponse, expected interface{}) {
    var actual interface{}
    err := resp.JSON(&actual)
    require.NoError(t, err)
    assert.Equal(t, expected, actual)
}

// 断言错误响应
func AssertErrorResponse(t *testing.T, resp *TestResponse, statusCode int, message string) {
    assert.Equal(t, statusCode, resp.StatusCode)
    
    var errorResp map[string]string
    resp.JSON(&errorResp)
    assert.Contains(t, errorResp["error"], message)
}

// 生成随机测试数据
func RandomString(length int) string {
    return gofakeit.LetterN(length)
}

func RandomEmail() string {
    return gofakeit.Email()
}

func RandomInt(min, max int) int {
    return gofakeit.Number(min, max)
}

// 时间测试助手
func TimeEqual(t *testing.T, expected, actual time.Time, delta time.Duration) {
    diff := actual.Sub(expected)
    if diff < 0 {
        diff = -diff
    }
    assert.True(t, diff <= delta, 
        "Time difference %v exceeds allowed delta %v", diff, delta)
}
```

YYHertz 的测试工具提供了完整的测试解决方案，从单元测试到性能测试，帮助开发者构建高质量、可靠的应用程序。通过合理使用这些工具，可以大大提高代码质量和开发效率。
