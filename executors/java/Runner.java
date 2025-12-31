import java.io.*;
import java.lang.reflect.*;
import java.nio.file.*;
import java.util.*;
import java.util.regex.*;
import javax.tools.*;

/**
 * Java code executor for Scribble.
 * Reads user code and test cases from environment variables,
 * compiles and executes the code against test cases, and returns results as JSON.
 */
public class Runner {

    private static final String CGROUP_MEMORY_PEAK = "/sys/fs/cgroup/memory.peak";
    private static final String CGROUP_MEMORY_EVENTS = "/sys/fs/cgroup/memory.events";

    public static void main(String[] args) {
        try {
            String codeB64 = System.getenv("CODE");
            String testCasesJson = System.getenv("TEST_CASES");
            String problemId = System.getenv("PROBLEM_ID");

            if (codeB64 == null || codeB64.isEmpty()) {
                printError("compilation_error", "No code provided", 0, 0);
                return;
            }

            // Decode code
            String code;
            try {
                code = new String(Base64.getDecoder().decode(codeB64), "UTF-8");
            } catch (Exception e) {
                printError("compilation_error", "Failed to decode code: " + e.getMessage(), 0, 0);
                return;
            }

            // Parse test cases
            List<Map<String, Object>> testCases;
            try {
                testCases = parseTestCases(testCasesJson);
            } catch (Exception e) {
                printError("runtime_error", "Failed to parse test cases: " + e.getMessage(), 0, 0);
                return;
            }

            long startMemory = measureMemory();

            // Compile code
            long compileStart = System.nanoTime();
            CompileResult compileResult = compileCode(code);
            long compileEnd = System.nanoTime();
            int compilationTimeMs = (int) ((compileEnd - compileStart) / 1_000_000);

            if (compileResult.error != null) {
                printError("compilation_error", compileResult.error, compilationTimeMs, 0);
                return;
            }

            // Find solution method
            Method solutionMethod = findSolutionMethod(compileResult.clazz);
            if (solutionMethod == null) {
                printError("compilation_error", "No solution method found", compilationTimeMs, 0);
                return;
            }

            // Execute tests
            List<Map<String, Object>> results = new ArrayList<>();
            int passedTests = 0;
            int totalExecutionTime = 0;
            boolean oomDetected = false;
            Object instance = compileResult.clazz.getDeclaredConstructor().newInstance();

            for (int i = 0; i < testCases.size(); i++) {
                Map<String, Object> testCase = testCases.get(i);
                Object input = testCase.get("input");
                Object expectedOutput = testCase.get("expected_output");

                long execStart = System.nanoTime();
                Object actualOutput = null;
                String error = null;

                try {
                    actualOutput = executeMethod(solutionMethod, instance, input);
                } catch (OutOfMemoryError e) {
                    oomDetected = true;
                    error = "Execution exceeded memory limit";
                } catch (Exception e) {
                    error = e.getMessage();
                }

                long execEnd = System.nanoTime();
                int execTimeMs = (int) ((execEnd - execStart) / 1_000_000);
                totalExecutionTime += execTimeMs;

                boolean passed = false;
                if (error == null) {
                    passed = compareOutputs(actualOutput, expectedOutput);
                    if (passed) passedTests++;
                }

                Map<String, Object> result = new LinkedHashMap<>();
                result.put("test_case_id", i);
                result.put("passed", passed);
                result.put("actual_output", error == null ? actualOutput : null);
                result.put("execution_time_ms", execTimeMs);
                result.put("error", error);
                results.add(result);
            }

            long peakMemory = measureMemory();
            int memoryUsedKb = (int) Math.max(peakMemory - startMemory, 0);

            // Check cgroups for accurate memory
            Long cgroupMemory = getCgroupsMemoryPeak();
            if (cgroupMemory != null) {
                memoryUsedKb = cgroupMemory.intValue();
            }

            int avgExecutionTime = testCases.size() > 0 ? totalExecutionTime / testCases.size() : 0;

            String status;
            if (oomDetected) {
                status = "memory_limit";
            } else if (passedTests == testCases.size()) {
                status = "accepted";
            } else if (results.stream().anyMatch(r -> r.get("error") != null)) {
                status = "runtime_error";
            } else {
                status = "wrong_answer";
            }

            printResult(status, passedTests, testCases.size(), compilationTimeMs,
                       avgExecutionTime, totalExecutionTime, memoryUsedKb, results);

        } catch (Exception e) {
            printError("runtime_error", "Executor error: " + e.getMessage(), 0, 0);
        }
    }

    private static Long getCgroupsMemoryPeak() {
        try {
            String content = new String(Files.readAllBytes(Paths.get(CGROUP_MEMORY_PEAK)));
            return Long.parseLong(content.trim()) / 1024; // bytes to KB
        } catch (Exception e) {
            return null;
        }
    }

    private static long measureMemory() {
        Runtime runtime = Runtime.getRuntime();
        return (runtime.totalMemory() - runtime.freeMemory()) / 1024; // KB
    }

    private static List<Map<String, Object>> parseTestCases(String json) throws Exception {
        // Simple JSON parser for test cases
        List<Map<String, Object>> result = new ArrayList<>();
        if (json == null || json.trim().isEmpty()) {
            return result;
        }

        json = json.trim();
        if (!json.startsWith("[")) {
            json = "[" + json + "]";
        }

        // Parse JSON array manually (simple implementation)
        int depth = 0;
        int start = -1;
        for (int i = 0; i < json.length(); i++) {
            char c = json.charAt(i);
            if (c == '{') {
                if (depth == 1) start = i;
                depth++;
            } else if (c == '}') {
                depth--;
                if (depth == 1 && start >= 0) {
                    String objStr = json.substring(start, i + 1);
                    result.add(parseJsonObject(objStr));
                    start = -1;
                }
            } else if (c == '[') {
                depth++;
            } else if (c == ']') {
                depth--;
            }
        }

        return result;
    }

    private static Map<String, Object> parseJsonObject(String json) throws Exception {
        Map<String, Object> result = new LinkedHashMap<>();
        json = json.trim();
        if (json.startsWith("{")) json = json.substring(1);
        if (json.endsWith("}")) json = json.substring(0, json.length() - 1);

        // Simple key-value parsing
        Pattern pattern = Pattern.compile("\"(\\w+)\"\\s*:\\s*(.+?)(?=,\\s*\"|$)");
        Matcher matcher = pattern.matcher(json);

        while (matcher.find()) {
            String key = matcher.group(1);
            String value = matcher.group(2).trim();
            result.put(key, parseJsonValue(value));
        }

        return result;
    }

    private static Object parseJsonValue(String value) {
        value = value.trim();
        if (value.endsWith(",")) value = value.substring(0, value.length() - 1).trim();

        if (value.equals("null")) return null;
        if (value.equals("true")) return true;
        if (value.equals("false")) return false;
        if (value.startsWith("\"") && value.endsWith("\"")) {
            return value.substring(1, value.length() - 1);
        }
        if (value.startsWith("[")) {
            return parseJsonArray(value);
        }
        try {
            if (value.contains(".")) return Double.parseDouble(value);
            return Long.parseLong(value);
        } catch (NumberFormatException e) {
            return value;
        }
    }

    private static List<Object> parseJsonArray(String json) {
        List<Object> result = new ArrayList<>();
        json = json.trim();
        if (json.startsWith("[")) json = json.substring(1);
        if (json.endsWith("]")) json = json.substring(0, json.length() - 1);

        if (json.trim().isEmpty()) return result;

        int depth = 0;
        StringBuilder current = new StringBuilder();
        for (char c : json.toCharArray()) {
            if (c == '[' || c == '{') depth++;
            else if (c == ']' || c == '}') depth--;

            if (c == ',' && depth == 0) {
                result.add(parseJsonValue(current.toString()));
                current = new StringBuilder();
            } else {
                current.append(c);
            }
        }
        if (current.length() > 0) {
            result.add(parseJsonValue(current.toString()));
        }

        return result;
    }

    private static class CompileResult {
        Class<?> clazz;
        String error;
    }

    private static CompileResult compileCode(String code) {
        CompileResult result = new CompileResult();

        try {
            // Extract class name
            Pattern pattern = Pattern.compile("class\\s+(\\w+)");
            Matcher matcher = pattern.matcher(code);
            String className = "Solution";
            if (matcher.find()) {
                className = matcher.group(1);
            }

            // Write source file
            Path sourceDir = Files.createTempDirectory("java_src");
            Path sourceFile = sourceDir.resolve(className + ".java");
            Files.write(sourceFile, code.getBytes());

            // Compile
            JavaCompiler compiler = ToolProvider.getSystemJavaCompiler();
            if (compiler == null) {
                result.error = "Java compiler not available";
                return result;
            }

            StringWriter errorWriter = new StringWriter();
            int compileStatus = compiler.run(null, null, errorWriter,
                sourceFile.toString(), "-d", sourceDir.toString());

            if (compileStatus != 0) {
                result.error = errorWriter.toString();
                return result;
            }

            // Load class
            java.net.URLClassLoader classLoader = new java.net.URLClassLoader(
                new java.net.URL[] { sourceDir.toUri().toURL() }
            );
            result.clazz = classLoader.loadClass(className);

        } catch (Exception e) {
            result.error = e.getMessage();
        }

        return result;
    }

    private static Method findSolutionMethod(Class<?> clazz) {
        String[] methodNames = {"solution", "solve", "main"};
        for (String name : methodNames) {
            for (Method m : clazz.getDeclaredMethods()) {
                if (m.getName().equals(name) && !Modifier.isPrivate(m.getModifiers())) {
                    return m;
                }
            }
        }
        // Return first public method
        for (Method m : clazz.getDeclaredMethods()) {
            if (Modifier.isPublic(m.getModifiers()) && !m.getName().equals("main")) {
                return m;
            }
        }
        return null;
    }

    private static Object executeMethod(Method method, Object instance, Object input) throws Exception {
        method.setAccessible(true);

        if (input instanceof List) {
            List<?> args = (List<?>) input;
            Object[] argsArray = args.toArray();
            // Convert Long to int if needed
            Class<?>[] paramTypes = method.getParameterTypes();
            for (int i = 0; i < argsArray.length && i < paramTypes.length; i++) {
                if (paramTypes[i] == int.class && argsArray[i] instanceof Long) {
                    argsArray[i] = ((Long) argsArray[i]).intValue();
                } else if (paramTypes[i] == int[].class && argsArray[i] instanceof List) {
                    List<?> list = (List<?>) argsArray[i];
                    int[] arr = new int[list.size()];
                    for (int j = 0; j < list.size(); j++) {
                        arr[j] = ((Number) list.get(j)).intValue();
                    }
                    argsArray[i] = arr;
                }
            }
            return method.invoke(instance, argsArray);
        }
        return method.invoke(instance, input);
    }

    private static boolean compareOutputs(Object actual, Object expected) {
        if (actual == null && expected == null) return true;
        if (actual == null || expected == null) return false;

        if (actual instanceof Number && expected instanceof Number) {
            double a = ((Number) actual).doubleValue();
            double e = ((Number) expected).doubleValue();
            return Math.abs(a - e) < 1e-9;
        }

        if (actual instanceof List && expected instanceof List) {
            List<?> aList = (List<?>) actual;
            List<?> eList = (List<?>) expected;
            if (aList.size() != eList.size()) return false;
            for (int i = 0; i < aList.size(); i++) {
                if (!compareOutputs(aList.get(i), eList.get(i))) return false;
            }
            return true;
        }

        if (actual.getClass().isArray() && expected instanceof List) {
            List<?> eList = (List<?>) expected;
            int len = java.lang.reflect.Array.getLength(actual);
            if (len != eList.size()) return false;
            for (int i = 0; i < len; i++) {
                if (!compareOutputs(java.lang.reflect.Array.get(actual, i), eList.get(i))) return false;
            }
            return true;
        }

        return actual.equals(expected);
    }

    private static void printError(String status, String message, int compilationTimeMs, int executionTimeMs) {
        System.out.println("{\"status\":\"" + status + "\",\"error_message\":\"" +
            escapeJson(message) + "\",\"compilation_time_ms\":" + compilationTimeMs +
            ",\"execution_time_ms\":" + executionTimeMs +
            ",\"memory_used_kb\":0,\"tests_passed\":0,\"tests_total\":0}");
    }

    private static void printResult(String status, int passed, int total,
            int compilationTimeMs, int avgExecTimeMs, int totalExecTimeMs,
            int memoryKb, List<Map<String, Object>> results) {
        StringBuilder sb = new StringBuilder();
        sb.append("{\"status\":\"").append(status).append("\"");
        sb.append(",\"tests_passed\":").append(passed);
        sb.append(",\"tests_total\":").append(total);
        sb.append(",\"compilation_time_ms\":").append(compilationTimeMs);
        sb.append(",\"execution_time_ms\":").append(avgExecTimeMs);
        sb.append(",\"total_execution_time_ms\":").append(totalExecTimeMs);
        sb.append(",\"memory_used_kb\":").append(memoryKb);
        sb.append(",\"test_results\":[");
        for (int i = 0; i < results.size(); i++) {
            if (i > 0) sb.append(",");
            sb.append(mapToJson(results.get(i)));
        }
        sb.append("]}");
        System.out.println(sb.toString());
    }

    private static String mapToJson(Map<String, Object> map) {
        StringBuilder sb = new StringBuilder("{");
        boolean first = true;
        for (Map.Entry<String, Object> entry : map.entrySet()) {
            if (!first) sb.append(",");
            first = false;
            sb.append("\"").append(entry.getKey()).append("\":");
            sb.append(valueToJson(entry.getValue()));
        }
        sb.append("}");
        return sb.toString();
    }

    private static String valueToJson(Object value) {
        if (value == null) return "null";
        if (value instanceof Boolean) return value.toString();
        if (value instanceof Number) return value.toString();
        if (value instanceof String) return "\"" + escapeJson((String) value) + "\"";
        if (value instanceof List) {
            StringBuilder sb = new StringBuilder("[");
            List<?> list = (List<?>) value;
            for (int i = 0; i < list.size(); i++) {
                if (i > 0) sb.append(",");
                sb.append(valueToJson(list.get(i)));
            }
            sb.append("]");
            return sb.toString();
        }
        return "\"" + escapeJson(value.toString()) + "\"";
    }

    private static String escapeJson(String s) {
        if (s == null) return "";
        return s.replace("\\", "\\\\")
                .replace("\"", "\\\"")
                .replace("\n", "\\n")
                .replace("\r", "\\r")
                .replace("\t", "\\t");
    }
}
