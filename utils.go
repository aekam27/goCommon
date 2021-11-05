package trestCommon

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	random "math/rand"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = random.New(random.NewSource(time.Now().UnixNano()))

func init() {
	LoadConfig()
}

func CreateToken(userid, email, name, status string) (string, error) {
	var err error
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["userid"] = userid
	atClaims["email"] = email
	atClaims["status"] = status
	if name != "" {
		atClaims["name"], _ = Decrypt(name)
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(viper.GetString("tokensecret")))
	if err != nil {
		ECLog3("unable to create token", err, logrus.Fields{"email": email, "user_id": userid, "name": name, "status": status})
		return "", err
	}
	return token, nil
}

func DecodeToken(tok string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tok, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(viper.GetString("tokensecret")), nil
	})
	if err != nil {
		ECLog3("unable to Decode token", err, logrus.Fields{"token": tok})
		return claims, err
	}
	return claims, nil
}

func ValidateEmail(email string) bool {
	var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if len(email) < 3 && len(email) > 254 {
		return false
	}
	return emailRegex.MatchString(email)
}

func PreSignedUrl(filename, path string) (string, error) {
	filename = strings.ReplaceAll(filename, " ", "")
	filename = strconv.Itoa(int(time.Now().Unix())) + filename
	// 	svc, err := createS3Session()
	// 	if err != nil {
	// 		ECLog2("unable to get s3 session", err)
	// 		return "", err
	// 	}
	// 	req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
	// 		Bucket: aws.String(viper.GetString("aws.bucket")),
	// 		Key:    aws.String(path + "/" + filename),
	// 	})
	// 	str, err := req.Presign(15 * time.Minute)
	// 	if err != nil {
	// 		ECLog2("failed to add expiry time to presigned url", err)
	// 		return "", err
	// 	}
	opts := &storage.SignedURLOptions{
		Scheme: storage.SigningSchemeV4,
		Method: "PUT",
		Headers: []string{
			"Content-Type:application/octet-stream",
		},
		GoogleAccessID: viper.GetString("gcp.email"),
		PrivateKey:     []byte(viper.GetString("gcp.private_key")),
		Expires:        time.Now().Add(15 * time.Minute),
	}
	str, err := storage.SignedURL(viper.GetString("gcp.bucket"), filename, opts)
	if err != nil {
		ECLog2("failed to create presigned url", err)
		return "", err
	}
	return str, nil
}
func PreSignedUrlAWS(filename, path string) (string, error) {
	filename = strings.ReplaceAll(filename, " ", "")
	filename = strconv.Itoa(int(time.Now().Unix())) + filename
	svc, err := createS3Session()
	if err != nil {
		ECLog2("unable to get s3 session", err)
		return "", err
	}
	req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(viper.GetString("aws.bucket")),
		Key:    aws.String(path + "/" + filename),
	})
	str, err := req.Presign(15 * time.Minute)
	if err != nil {
		ECLog2("failed to add expiry time to presigned url", err)
		return "", err
	}
	return str, nil
}
func PreSignedDownloadUrlAWS(filename, path string) (string, error) {
	svc, err := createS3Session()
	if err != nil {
		ECLog2("unable to get s3 session", err)
		return "", err
	}
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(viper.GetString("aws.bucket")),
		Key:    aws.String(path + "/" + filename),
	})
	str, err := req.Presign(15 * time.Minute)
	if err != nil {
		ECLog2("failed to add expiry time to presigned url", err)
		return "", err
	}
	return str, nil
}

func Encrypt(text string) (string, error) {
	key := []byte(viper.GetString("encryptionkey"))
	plaintext := []byte(text)
	block, err := aes.NewCipher(key)
	if err != nil {
		ECLog3("unable to create block", err, logrus.Fields{"key": viper.GetString("encryptionkey"), "text": text})
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		ECLog3("unable to create aesGCM", err, logrus.Fields{"key": viper.GetString("encryptionkey"), "text": text})
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		ECLog3("encrypt unable to read", err, logrus.Fields{"key": viper.GetString("encryptionkey"), "text": text})
		return "", err
	}
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return fmt.Sprintf("%x", ciphertext), nil
}

func Decrypt(text string) (string, error) {
	key := []byte(viper.GetString("encryptionkey"))
	enc, _ := hex.DecodeString(text)
	block, err := aes.NewCipher(key)
	if err != nil {
		ECLog3("unable to create block", err, logrus.Fields{"key": viper.GetString("encryptionkey"), "text": text})
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		ECLog3("unable to create aesGCM", err, logrus.Fields{"key": viper.GetString("encryptionkey"), "text": text})
		return "", err
	}
	nonceSize := aesGCM.NonceSize()
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		ECLog3("unable to decode", err, logrus.Fields{"key": viper.GetString("encryptionkey"), "text": text})
		return "", err
	}
	return fmt.Sprintf("%s", plaintext), nil
}

func GetRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func SendVerificationCode(email, verificationCode string) (string, error) {
	url := createUrl(verificationCode, "verifyemail")
	subject := "Verification Email"
	htmlBody := "Hi " + email + ",<br><br>"+viper.GetString("email.template")+"<br><br><a href=" + url + ">Verification Link</a>"
	textBody := "Hi " + email + ",\n\n"+ viper.GetString("email.template") + url + ""
	return sendEmail(email, subject, htmlBody, textBody)
}

func SendResetPasswordLink(email, resetCode string) (string, error) {
	url := createUrl(resetCode, "resetpassword")
	subject := "Reset Password"
	htmlBody := "Hi " + email + ",<br><br>Please click on the below url to reset your account password<br><br><a href=" + url + ">Reset Password Link</a>"
	textBody := "Hi " + email + ",\n\nPlease click on the below url to reset your account password" + url + ""
	return sendEmail(email, subject, htmlBody, textBody)
}

func sendEmail(email, subject, htmlBody, textBody string) (string, error) {
	svc, err := createSeSSession()
	if err != nil {
		ECLog3("send email verification failed", err, logrus.Fields{"email": email, "htmlBody": htmlBody})
		return "", err
	}
	from := viper.GetString("email.from")
	to := email
	input := &ses.SendEmailInput{
		Source: &from,
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Data: aws.String(htmlBody),
				},
				Text: &ses.Content{
					Data: aws.String(textBody),
				},
			},
			Subject: &ses.Content{
				Data: aws.String(subject),
			},
		},
		Destination: &ses.Destination{
			ToAddresses: []*string{&to},
		},
	}
	_, err = svc.SendEmail(input)
	if err != nil {
		ECLog3("send email verification failed", err, logrus.Fields{"email": email, "htmlBody": htmlBody})
		return "", err
	}
	return "Sent Successfully", nil
}

func createUrl(verificationcode, path string) string {
	cart := viper.GetString("website.url")
	website := cart
	if strings.Contains(cart, "https") {
		cartSplit := strings.Split(cart, "/")
		website = cartSplit[2]
	}
	u := &url.URL{
		Scheme: "https",
		Host:   website,
		Path:   path + "/" + verificationcode,
	}
	return u.String()
}

func GetHeader(uRL string, limit int) ([][]string, error) {
	_, err := url.ParseRequestURI(uRL)
	if err != nil {
		return [][]string{}, err
	}
	urldata := strings.Split(uRL, "/")
	bucket := strings.Split(urldata[2], ".")[0]
	path := urldata[3]
	filename := urldata[4]
	if strings.Contains(urldata[4], "?") {
		filename = strings.Split(urldata[4], "?")[0]
	}
	dataQuery := ""
	if limit == 0 {
		dataQuery = "SELECT s.* FROM S3Object s"
	} else {
		dataQuery = "SELECT s.* FROM S3Object s limit " + strconv.Itoa(limit)
	}
	data, err := getData(bucket, filename, path, dataQuery)
	if err != nil {
		ECLog3("getting header failed", err, logrus.Fields{"url": uRL})
		return [][]string{}, err
	}
	return data, nil
}

func getData(bucket, filename, path, Query string) ([][]string, error) {
	svc, err := createS3Session()
	if err != nil {
		ECLog3("getting data failed", err, logrus.Fields{"bucket": bucket, "filename": filename, "path": path})
		return [][]string{}, err
	}
	params := getS3SQLQueryParameters(bucket, filename, path, Query)
	resp, err := svc.SelectObjectContent(params)
	if err != nil {
		ECLog3("getting data failed", err, logrus.Fields{"bucket": bucket, "filename": filename, "path": path})
		return [][]string{}, err
	}
	defer resp.EventStream.Close()
	results, resultWriter := io.Pipe()
	go func() {
		defer resultWriter.Close()
		for event := range resp.EventStream.Events() {
			switch e := event.(type) {
			case *s3.RecordsEvent:
				resultWriter.Write(e.Payload)
			}
		}
	}()
	return readCSVFile(results)
}

func getS3SQLQueryParameters(bucket, filename, path, sqlquery string) *s3.SelectObjectContentInput {
	params := &s3.SelectObjectContentInput{
		Bucket:         aws.String(bucket),
		Key:            aws.String(path + "/" + filename),
		ExpressionType: aws.String(s3.ExpressionTypeSql),
		Expression:     aws.String(sqlquery),
		InputSerialization: &s3.InputSerialization{
			CSV: &s3.CSVInput{
				FileHeaderInfo: aws.String(s3.FileHeaderInfoNone),
			},
		},
		OutputSerialization: &s3.OutputSerialization{
			CSV: &s3.CSVOutput{},
		},
	}
	return params
}

func readCSVFile(results *io.PipeReader) ([][]string, error) {
	resReader := csv.NewReader(results)
	csvData, err := resReader.ReadAll()
	if err != nil {
		ECLog2("reading scv files failed", err)
		return [][]string{}, err
	}
	return formatCSV(csvData)
}

func formatCSV(csvdata [][]string) ([][]string, error) {
	headers := csvdata[0]
	data := [][]string{}
	for i := range headers {
		temp := []string{headers[i]}
		for j := 1; j < len(csvdata); j++ {
			temp = append(temp, csvdata[j][i])
		}
		data = append(data, temp)
	}
	return data, nil
}

func createS3Session() (*s3.S3, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(viper.GetString("aws.region")),
		Credentials: credentials.NewStaticCredentials(viper.GetString("aws.aws_access_key_id"),
			viper.GetString("aws.aws_secret_access_key"), "")},
	)
	if err != nil {
		ECLog2("creating s3 session", err)
		return nil, err
	}
	svc := s3.New(sess)
	return svc, nil
}

func createSeSSession() (*ses.SES, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(viper.GetString("aws.region")),
		Credentials: credentials.NewStaticCredentials(viper.GetString("aws.aws_access_key_id"),
			viper.GetString("aws.aws_secret_access_key"), "")},
	)
	if err != nil {
		ECLog2("creating ses session", err)
		return nil, err
	}
	svc := ses.New(sess)
	return svc, nil
}
