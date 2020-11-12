using System;
using Audit.Core;
using Newtonsoft.Json;

namespace TryCustomAuditNet
{
    class Program
    {
        static void Main(string[] args)
        {
            ConfigureAudit();
            
            var order = new Order(Guid.NewGuid(), "Jone Doe", 100, DateTime.UtcNow);
            using (var scope = AuditScope.Create("Order::Update", () => order))
            {
                order.UpdateOrderAmount(200);
                
                // optional
                scope.Comment("this is a test for update order.");
            }
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