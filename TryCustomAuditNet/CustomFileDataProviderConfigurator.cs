using System;
using Audit.Core;
using Audit.Core.ConfigurationApi;
using Newtonsoft.Json;

namespace TryCustomAuditNet
{
    public class CustomFileDataProviderConfigurator: IFileLogProviderConfigurator
    {
        public JsonSerializerSettings _jsonSettings = null;
        public string _directoryPath = "";
        public string _filenamePrefix = ""; 
        public Func<AuditEvent, string> _filenameBuilder; 
        public Func<AuditEvent, string> _directoryPathBuilder;

        public IFileLogProviderConfigurator JsonSettings(JsonSerializerSettings jsonSettings)
        {
            _jsonSettings = jsonSettings;
            return this;
        }

        public IFileLogProviderConfigurator Directory(string directoryPath)
        {
            _directoryPath = directoryPath;
            return this;
        }

        public IFileLogProviderConfigurator DirectoryBuilder(Func<AuditEvent, string> directoryPathBuilder)
        {
            _directoryPathBuilder = directoryPathBuilder;
            return this;
        }

        public IFileLogProviderConfigurator FilenamePrefix(string filenamePrefix)
        {
            _filenamePrefix = filenamePrefix;
            return this;
        }

        public IFileLogProviderConfigurator FilenameBuilder(Func<AuditEvent, string> filenameBuilder)
        {
            _filenameBuilder = filenameBuilder;
            return this;
        }
    }
}