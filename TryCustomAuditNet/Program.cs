using System;
using Amazon.Runtime;
using Amazon.SQS;
using Audit.Core;
using Newtonsoft.Json;

namespace TryCustomAuditNet
{
    class Program
    {
        const string AWS_ACCESS_KEY_ID = "awskey";
        const string AWS_SECRET_ACCESS_KEY = "awssecret";
        
        static void Main(string[] args)
        {
            var queueUrl = "http://localhost:4576/queue/audit";
            
            var client = BuildAmazonSqsClient("http://localhost:4576");
            
            ConfigureAudit();
            
            var order = new Order("BASE_name", Guid.NewGuid(), "Jone Doe", 100, DateTime.UtcNow, new Product
            {
                ProductId = Guid.NewGuid(),
                ProductName = "Test111111",
                ProductPrice = 30
            });
            using (var scope = AuditScope.Create("Order::Update", () => order))
            {
                order.UpdateOrderAmount(200);
                order.UpdateName(null);
                
                // optional
                scope.Comment("this is a test for update order.");
            }
        }

        private static void ConfigureAudit(AmazonSQSClient client, string queueUrl)
        {
            Audit.Core.Configuration.Setup()
                .UseCustomProvider(new AmazonSqsDataProvider(client, queueUrl));
        }

        private static AmazonSQSClient BuildAmazonSqsClient(string queueUrl)
        {
            var sqsConfig = new AmazonSQSConfig
            {
                ServiceURL = "http://localhost:4576",
                UseHttp = true,
                AuthenticationRegion = "ap-southeast-2",
            };
            
            AWSCredentials creds = new BasicAWSCredentials(AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY);
            return new AmazonSQSClient(creds, sqsConfig);
        }

        private static void ConfigureAudit()
        {
            Audit.Core.Configuration.Setup()
                .UseCustomProvider(new CustomFileDataProvider(config => config
                    .DirectoryBuilder(_ => "./")
                    .FilenameBuilder(auditEvent => $"{auditEvent.EventType}_{DateTime.Now.Ticks}.json")
                    .JsonSettings(new JsonSerializerSettings
                    {
                        Formatting = Formatting.Indented,
                        ReferenceLoopHandling = ReferenceLoopHandling.Ignore,
                        NullValueHandling = NullValueHandling.Include
                    })));
        }
    }
}