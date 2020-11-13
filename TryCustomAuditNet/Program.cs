using System;
using Amazon;
using Amazon.Runtime;
using Amazon.SQS;
using Audit.Core;
using Newtonsoft.Json;

namespace TryCustomAuditNet
{
    class Program
    {
        static void Main(string[] args)
        {
            var queueUrl = "http://localstack:4576/queue/audit";
            var client = BuildAmazonSqsClient("http://localstack:4576");

            ConfigureAudit(client, queueUrl);
            
            var order = new Order(Guid.NewGuid(), "Jone Doe", 100, DateTime.UtcNow);
            using (var scope = AuditScope.Create("Order::Update", () => order))
            {
                order.UpdateOrderAmount(200);
                
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
            var config = new AmazonSQSConfig();
            config.ServiceURL = queueUrl;
            config.RegionEndpoint = RegionEndpoint.APSoutheast2;
            return new AmazonSQSClient(config);
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