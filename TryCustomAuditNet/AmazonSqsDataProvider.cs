using System;
using System.Reflection;
using System.Threading;
using System.Threading.Tasks;
using Amazon.SQS;
using Amazon.SQS.Model;
using Audit.Core;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;

namespace TryCustomAuditNet
{
    public class AmazonSqsDataProvider: AuditDataProvider
    {
        private readonly IAmazonSQS _amazonSqs;
        private readonly string _queueUrl;
        
        public AmazonSqsDataProvider(IAmazonSQS amazonSqs, string queueUrl)
        {
            _amazonSqs = amazonSqs;
            _queueUrl = queueUrl;
        }
        
        public override object Serialize<T>(T value)
        {
            if (value == null)
            {
                return null;
            }
            
            var jo = new JObject();
            var serializer = JsonSerializer.Create(Configuration.JsonSettings);
            
            foreach (PropertyInfo propInfo in value.GetType().GetProperties())
            {
                if (propInfo.CanRead)
                {
                    object propVal = propInfo.GetValue(value, null);
                        
                    var customAttribute = propInfo.GetCustomAttribute<UnAuditableAttribute>();
                    if (customAttribute == null)
                    {
                        if (propVal == null)
                        {
                            jo.Add(propInfo.Name, JValue.CreateNull());
                        }
                        else
                        {
                            jo.Add(propInfo.Name, JToken.FromObject(propVal, serializer));
                        }
                    }
                }
            }
            
            return JToken.FromObject(jo, serializer);
        }

        public override object InsertEvent(AuditEvent auditEvent)
        {
            SendMessageToSqs(_queueUrl, auditEvent).GetAwaiter().GetResult();
            return null;
        }

        private async Task SendMessageToSqs(string queueUrl, AuditEvent auditEvent, CancellationToken cancellationToken = default(CancellationToken))
        {
            if(string.IsNullOrWhiteSpace(queueUrl)) return;
            
            var message = JsonConvert.SerializeObject(auditEvent, Formatting.None);
            Console.WriteLine($"send message is {message} to {queueUrl}");

            var request = new SendMessageRequest(queueUrl, message)
            {
                MessageBody = message,
                QueueUrl = queueUrl
            };
            
            await _amazonSqs.SendMessageAsync(request, cancellationToken);
        }
    }
}