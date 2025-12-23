#!/usr/bin/env ruby
# frozen_string_literal: true

# ELK Helper 测试日志生成脚本
# 用法: ruby scripts/generate-test-logs.rb [options]
#
# 默认配置（可通过环境变量覆盖）:
#   ES_URL       - Elasticsearch 地址 (默认: https://localhost:9200)
#   ES_USERNAME  - ES 用户名 (默认: elastic)
#   ES_PASSWORD  - ES 密码 (默认: changeme)
#   ES_SSL       - 是否使用 SSL (自动检测，默认: true for https://)
#   ES_SKIP_VERIFY - 是否跳过证书验证 (默认: true，方便测试)
#
# 示例:
#   ruby scripts/generate-test-logs.rb --type nginx --count 1000
#   ruby scripts/generate-test-logs.rb --type java --count 500 --prefix app-logs
#   ruby scripts/generate-test-logs.rb --type both --count 2000 --days 7
#
# 自定义 ES 连接（使用环境变量）:
#   export ES_URL="https://es.example.com:9200"
#   export ES_USERNAME="myuser"
#   export ES_PASSWORD="mypass"
#   ruby scripts/generate-test-logs.rb --type nginx --count 1000

require 'optparse'
require 'net/http'
require 'uri'
require 'json'
require 'securerandom'
require 'date'

class ESTestLogGenerator
  DEFAULT_CONFIG = {
    url: ENV['ES_URL'] || 'https://localhost:9200',
    username: ENV['ES_USERNAME'] || 'elastic',
    password: ENV['ES_PASSWORD'] || 'changeme',
    ssl: nil,  # 自动从 URL 检测
    skip_verify: nil  # 自动设置为 true（如果使用 HTTPS）
  }.freeze

  HTTP_CODES = [200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200,
                200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200,
                301, 302, 304, 400, 401, 403, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404,
                500, 500, 500, 502, 502, 503, 503, 504, 499].freeze

  USER_AGENTS = [
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
    'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36',
    'Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15',
    'Mozilla/5.0 (Android 10; Mobile; rv:68.0) Gecko/68.0',
    'curl/7.68.0',
    'PostmanRuntime/7.26.8',
    'blackbox-monitoring'
  ].freeze

  REQUESTS = [
    '/api/v1/users',
    '/api/v1/orders',
    '/api/v1/products',
    '/api/v1/payments',
    '/api/v1/auth/login',
    '/api/v1/auth/logout',
    '/api/v1/config',
    '/health',
    '/status',
    '/index.html',
    '/static/css/main.css',
    '/static/js/app.js',
    '/favicon.ico',
    '/aggamepch5/config.js',
    '/api/v1/data',
    '/api/v1/search',
    '/api/v1/export',
    '/dashboard',
    '/login',
    '/logout'
  ].freeze

  JAVA_MESSAGES = [
    'invoking api:DomainConfigOperationService.queryDomains',
    'User login successful',
    'Database connection established',
    'Cache updated successfully',
    'File uploaded: document.pdf',
    'Payment processed: order_id=12345',
    'Exception in thread "main" java.lang.NullPointerException',
    'SQLException: Connection timeout',
    'ERROR: Failed to process request',
    'WARN: High memory usage detected',
    'INFO: Application started successfully',
    'DEBUG: Processing request with id: 12345',
    'ERROR: Database query failed',
    'Exception: java.net.SocketTimeoutException',
    'SQLException: Deadlock detected',
    'ERROR: OutOfMemoryError in heap space'
  ].freeze

  MODULES = ['g01-gci-api', 'g01-gci-web', 'g01-gci-admin', 'g01-gci-payment', 'g01-gci-notification'].freeze
  PROJECTS = ['g01-prod', 'g01-staging', 'g01-dev'].freeze
  ENVS = ['prod', 'staging', 'dev'].freeze
  HOSTNAMES = ['gcp-hk-g01-gamepublic-nginx-01', 'gcp-hk-g01-gamepublic-nginx-02', 'gcp-hk-g01-gamepublic-nginx-03',
               'gcp-hk-g01-api-server-01', 'gcp-hk-g01-api-server-02'].freeze
  IPS = ['103.250.7.186', '103.250.7.187', '103.250.7.188', '192.168.1.100', '192.168.1.101'].freeze
  DOMAINS = ['gci-web.*', 'api.example.com', 'admin.example.com', 'web.example.com'].freeze

  def initialize(options = {})
    @config = DEFAULT_CONFIG.merge(options)
    # Auto-detect SSL from URL if not explicitly set via ES_SSL env var
    if ENV['ES_SSL'].nil? && @config[:url]
      urls = @config[:url].split(';').map(&:strip).reject(&:empty?)
      @config[:ssl] = urls.any? { |u| u.downcase.start_with?('https://') }
    end
    # Auto-enable skip_verify for HTTPS (default is true for testing convenience)
    if @config[:ssl] && ENV['ES_SKIP_VERIFY'].nil?
      @config[:skip_verify] = true
    end
    @type = options[:type] || 'both'
    @count = options[:count] || 1000
    @index_prefix = options[:index_prefix] || 'test'
    @days = options[:days] || 1
    @batch_size = options[:batch_size] || 100
  end

  def run
    validate_config!
    print_header

    # 解析 ES URL
    urls = parse_es_urls
    puts "连接到 Elasticsearch: #{urls.size} 个节点"
    puts

    # 测试连接
    test_connection(urls.first)

    # 生成日志
    case @type
    when 'nginx'
      generate_nginx_logs(urls)
    when 'java'
      generate_java_logs(urls)
    when 'both'
      generate_nginx_logs(urls)
      generate_java_logs(urls)
    else
      puts "错误: 未知的日志类型: #{@type}"
      puts "支持的类型: nginx, java, both"
      exit 1
    end

    puts
    puts "✅ 完成！"
  end

  private

  def validate_config!
    if @config[:url].nil? || @config[:url].empty?
      puts "错误: ES_URL 未设置"
      exit 1
    end
  end

  def print_header
    puts '=' * 60
    puts 'ELK Helper 测试日志生成工具'
    puts '=' * 60
    puts
    puts "配置信息:"
    puts "  ES 地址: #{@config[:url]}"
    puts "  日志类型: #{@type}"
    puts "  生成数量: #{@count}"
    puts "  索引前缀: #{@index_prefix}"
    puts "  时间范围: 最近 #{@days} 天"
    puts "  批量大小: #{@batch_size}"
    puts
  end

  def parse_es_urls
    @config[:url].split(';').map(&:strip).reject(&:empty?)
  end

  def test_connection(url)
    uri = URI.parse(url)
    http = create_http(uri)
    request = Net::HTTP::Get.new('/')
    request.basic_auth(@config[:username], @config[:password]) if @config[:username]

    response = http.request(request)
    if response.code == '200'
      puts "✅ ES 连接成功"
    else
      puts "⚠️  ES 连接返回: #{response.code}"
    end
  rescue StandardError => e
    puts "❌ ES 连接失败: #{e.message}"
    exit 1
  end

  def create_http(uri)
    http = Net::HTTP.new(uri.host, uri.port)
    # Auto-detect SSL from URL scheme, or use explicit SSL config
    if uri.scheme == 'https' || @config[:ssl]
      http.use_ssl = true
      # Skip certificate verification if explicitly set, or if URL is https:// (default to skip for convenience)
      if @config[:skip_verify] || (uri.scheme == 'https' && !@config.key?(:skip_verify))
        http.verify_mode = OpenSSL::SSL::VERIFY_NONE
      end
    end
    http.read_timeout = 30
    http.open_timeout = 10
    http
  end

  def generate_nginx_logs(urls)
    puts "生成 Nginx 日志..."
    index_name = "#{@index_prefix}-nginx-access"
    generate_logs(urls, index_name, method(:generate_nginx_log_entry))
  end

  def generate_java_logs(urls)
    puts "生成 Java 日志..."
    index_name = "#{@index_prefix}-java"
    generate_logs(urls, index_name, method(:generate_java_log_entry))
  end

  def generate_logs(urls, index_base, entry_generator)
    total = 0
    url_index = 0

    # 生成多天的索引
    @days.times do |day_offset|
      date = Date.today - day_offset
      index_name = "#{index_base}-#{date.strftime('%Y.%m.%d')}"

      puts "  创建索引: #{index_name}"

      # 创建索引
      create_index(urls.first, index_name)

      # 生成日志
      batch = []
      @count.times do |i|
        timestamp = generate_timestamp(day_offset)
        entry = entry_generator.call(timestamp)
        batch << { index: { _index: index_name } }.to_json
        batch << entry.to_json

        if batch.size >= @batch_size * 2
          bulk_index(urls[url_index % urls.size], batch)
          total += batch.size / 2
          batch.clear
          print "\r    已生成: #{total} 条" if (i % 100).zero?
          url_index += 1
        end
      end

      # 处理剩余批次
      unless batch.empty?
        bulk_index(urls[url_index % urls.size], batch)
        total += batch.size / 2
        url_index += 1
      end

      puts "\r    已生成: #{total} 条日志到索引 #{index_name}"
    end

    puts "  ✅ 总共生成 #{total} 条日志"
  end

  def generate_nginx_log_entry(timestamp)
    response_code = HTTP_CODES.sample
    responsetime = rand(0.001..5.0).round(3)
    request = REQUESTS.sample
    ip = IPS.sample
    hostname = HOSTNAMES.sample
    user_agent = USER_AGENTS.sample
    domain = DOMAINS.sample

    {
      '@timestamp' => timestamp.iso8601(3),
      '@version' => '1',
      'log_type' => 'nginx-access',
      'scheme' => ['http', 'https'].sample,
      'hostname' => hostname,
      'ip' => ip,
      'request_method' => ['GET', 'POST', 'PUT', 'DELETE', 'PATCH'].sample,
      'request' => request,
      'response_code' => response_code,
      'responsetime' => responsetime,
      'size' => rand(100..10000),
      'http_user_agent' => user_agent,
      'referer' => ['-', 'https://example.com', 'https://google.com'].sample,
      'domain' => domain,
      'project' => PROJECTS.sample,
      'module' => 'g01-gci-web-access',
      'env' => ENVS.sample,
      'node_ip' => "10.170.96.#{rand(1..255)}",
      'upstreamaddr' => "10.170.96.#{rand(1..255)}:8888",
      'upstreamtime' => rand(0.0..2.0).round(3),
      'cf_ray' => ['-', SecureRandom.hex(8)].sample,
      'cf_ipcountry' => ['-', 'CN', 'US', 'HK', 'SG'].sample,
      'https' => ['', 'on'].sample,
      'args' => ['-', 'id=123'].sample,
      'geo' => {
        'continent_code' => 'AS',
        'country_code' => ['CN', 'US', 'HK', 'SG', 'MY'].sample,
        'coordinates' => ["#{rand(100..120)}.#{rand(0..9)}", "#{rand(20..40)}.#{rand(0..9)}"],
        'timezone' => ['Asia/Shanghai', 'Asia/Hong_Kong', 'UTC'].sample
      },
      'ua' => {
        'name' => 'Chrome',
        'os' => 'Windows',
        'os_name' => 'Windows',
        'device' => 'Desktop',
        'os_full' => 'Windows 10'
      },
      'fields' => {},
      'input' => { 'type' => 'log' }
    }
  end

  def generate_java_log_entry(timestamp)
    message = JAVA_MESSAGES.sample
    module_name = MODULES.sample
    project = PROJECTS.sample
    env = ENVS.sample
    node_ip = "10.170.96.#{rand(1..255)}"
    timestamp2 = timestamp.strftime('%Y-%m-%d %H:%M:%S.%L')

    {
      '@timestamp' => timestamp.iso8601(3),
      '@version' => '1',
      'log_type' => 'java',
      'message' => message,
      'timestamp2' => timestamp2,
      'module' => module_name,
      'project' => project,
      'env' => env,
      'node_ip' => node_ip,
      'event' => {},
      'fields' => {},
      'input' => { 'type' => 'log' }
    }
  end

  def generate_timestamp(day_offset)
    now = Time.now
    base_time = now - (day_offset * 24 * 60 * 60)
    # 随机时间，分布在最近 24 小时内
    random_seconds = rand(0..(24 * 60 * 60))
    Time.at(base_time.to_i - random_seconds)
  end

  def create_index(url, index_name)
    uri = URI.parse("#{url}/#{index_name}")
    http = create_http(uri)
    request = Net::HTTP::Put.new(uri.path)
    request['Content-Type'] = 'application/json'
    request.basic_auth(@config[:username], @config[:password]) if @config[:username]

    # 简单的索引映射
    mapping = {
      settings: {
        number_of_shards: 1,
        number_of_replicas: 0
      },
      mappings: {
        properties: {
          '@timestamp' => { type: 'date' },
          'response_code' => { type: 'integer' },
          'responsetime' => { type: 'float' },
          'log_type' => { type: 'keyword' }
        }
      }
    }
    request.body = mapping.to_json

    response = http.request(request)
    return if response.code == '200' || response.code == '201'

    # 如果索引已存在，忽略错误
    return if response.body.include?('resource_already_exists_exception')

    puts "  ⚠️  创建索引失败: #{response.code} - #{response.body[0..100]}"
  rescue StandardError => e
    puts "  ⚠️  创建索引异常: #{e.message}"
  end

  def bulk_index(url, batch)
    uri = URI.parse("#{url}/_bulk")
    http = create_http(uri)
    request = Net::HTTP::Post.new(uri.path)
    request['Content-Type'] = 'application/x-ndjson'
    request.basic_auth(@config[:username], @config[:password]) if @config[:username]
    request.body = batch.join("\n") + "\n"

    response = http.request(request)
    unless response.code == '200' || response.code == '201'
      puts "\n  ⚠️  批量插入失败: #{response.code}"
      puts "  #{response.body[0..200]}"
    end
  rescue StandardError => e
    puts "\n  ⚠️  批量插入异常: #{e.message}"
  end
end

# 命令行参数解析
options = {}
OptionParser.new do |opts|
  opts.banner = "用法: ruby #{$PROGRAM_NAME} [options]"

  opts.on('-t', '--type TYPE', '日志类型: nginx, java, both (默认: both)') do |t|
    options[:type] = t
  end

  opts.on('-c', '--count COUNT', Integer, '生成数量 (默认: 1000)') do |c|
    options[:count] = c
  end

  opts.on('-p', '--prefix PREFIX', '索引前缀 (默认: test)') do |p|
    options[:index_prefix] = p
  end

  opts.on('-d', '--days DAYS', Integer, '生成最近几天的数据 (默认: 1)') do |d|
    options[:days] = d
  end

  opts.on('-b', '--batch-size SIZE', Integer, '批量大小 (默认: 100)') do |b|
    options[:batch_size] = b
  end

  opts.on('-h', '--help', '显示帮助信息') do
    puts opts
    exit
  end
end.parse!

# 运行生成器
generator = ESTestLogGenerator.new(options)
generator.run

